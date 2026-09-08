import torch
import triton
import triton.language as tl


@triton.jit
def _kv_rmsnorm_rope_cache_norm_kernel(
    kv_ptr,
    gamma_ptr,
    cos_ptr,
    sin_ptr,
    index_ptr,
    k_cache_ptr,
    ckv_cache_ptr,
    k_scale_ptr,
    c_scale_ptr,
    k_offset_ptr,
    c_offset_ptr,
    stride_kv_b: tl.constexpr,
    stride_kv_h: tl.constexpr,
    stride_kv_t: tl.constexpr,
    stride_cos_b: tl.constexpr,
    stride_cos_h: tl.constexpr,
    stride_cos_t: tl.constexpr,
    stride_sin_b: tl.constexpr,
    stride_sin_h: tl.constexpr,
    stride_sin_t: tl.constexpr,
    stride_index_b: tl.constexpr,
    stride_k_b: tl.constexpr,
    stride_k_h: tl.constexpr,
    stride_k_s: tl.constexpr,
    stride_c_b: tl.constexpr,
    stride_c_h: tl.constexpr,
    stride_c_s: tl.constexpr,
    cache_len: tl.constexpr,
    seq_len: tl.constexpr,
    num_heads: tl.constexpr,
    epsilon: tl.constexpr,
    has_k_scale: tl.constexpr,
    has_c_scale: tl.constexpr,
    has_k_offset: tl.constexpr,
    has_c_offset: tl.constexpr,
    block_t: tl.constexpr,
):
    token = tl.program_id(0)
    batch_head = tl.program_id(1)
    batch = batch_head // num_heads
    head = batch_head - batch * num_heads

    token_ids = tl.arange(0, block_t)
    valid_tokens = token_ids < seq_len
    index_sequence = index_ptr + batch * stride_index_b
    indices = tl.load(index_sequence + token_ids, mask=valid_tokens, other=-1).to(tl.int32)
    position = tl.load(index_sequence + token).to(tl.int32)
    winners = tl.where(valid_tokens & (indices == position), token_ids, -1)
    last_token = tl.max(winners, axis=0)
    write_mask = (token == last_token) & (position >= 0) & (position < cache_len)

    c_cols = tl.arange(0, 512)
    r_cols = tl.arange(0, 64)
    half_cols = r_cols % 32
    rope_real_col = 512 + half_cols * 2
    rope_imag_col = rope_real_col + 1
    rope_real_scale_col = half_cols * 2
    rope_imag_scale_col = rope_real_scale_col + 1

    gamma = tl.load(gamma_ptr + c_cols).to(tl.float32)
    if has_c_scale:
        c_scale = tl.load(c_scale_ptr + c_cols).to(tl.float32)
    if has_c_offset:
        c_offset = tl.load(c_offset_ptr + c_cols).to(tl.float32)

    kv_sequence = kv_ptr + batch * stride_kv_b + head * stride_kv_h
    cos_sequence = cos_ptr + batch * stride_cos_b + head * stride_cos_h
    sin_sequence = sin_ptr + batch * stride_sin_b + head * stride_sin_h
    kv_row = kv_sequence + token * stride_kv_t

    ckv = tl.load(kv_row + c_cols).to(tl.float32)
    square_sum = tl.sum(ckv * ckv, axis=0)
    rms = tl.sqrt(square_sum * (1.0 / 512.0) + epsilon)
    ckv_out = ckv / rms * gamma
    if has_c_scale:
        ckv_out = ckv_out * c_scale
    if has_c_offset:
        ckv_out = ckv_out + c_offset

    real = tl.load(kv_row + rope_real_col).to(tl.float32)
    imag = tl.load(kv_row + rope_imag_col).to(tl.float32)
    if has_k_scale:
        real = real * tl.load(k_scale_ptr + rope_real_scale_col).to(tl.float32)
        imag = imag * tl.load(k_scale_ptr + rope_imag_scale_col).to(tl.float32)
    if has_k_offset:
        real = real + tl.load(k_offset_ptr + rope_real_scale_col).to(tl.float32)
        imag = imag + tl.load(k_offset_ptr + rope_imag_scale_col).to(tl.float32)

    cos_value = tl.load(cos_sequence + token * stride_cos_t + r_cols).to(tl.float32)
    sin_value = tl.load(sin_sequence + token * stride_sin_t + r_cols).to(tl.float32)
    rope_out = tl.where(
        r_cols < 32,
        real * cos_value - imag * sin_value,
        imag * cos_value + real * sin_value,
    )

    k_dst = (
        k_cache_ptr
        + batch * stride_k_b
        + head * stride_k_h
        + position * stride_k_s
        + r_cols
    )
    c_dst = (
        ckv_cache_ptr
        + batch * stride_c_b
        + head * stride_c_h
        + position * stride_c_s
        + c_cols
    )
    tl.store(k_dst, rope_out, mask=write_mask)
    tl.store(c_dst, ckv_out, mask=write_mask)


@triton.jit
def _kv_rmsnorm_rope_cache_kernel(
    kv_ptr,
    gamma_ptr,
    cos_ptr,
    sin_ptr,
    index_ptr,
    k_cache_ptr,
    ckv_cache_ptr,
    k_scale_ptr,
    c_scale_ptr,
    k_offset_ptr,
    c_offset_ptr,
    stride_kv_b: tl.constexpr,
    stride_kv_h: tl.constexpr,
    stride_kv_t: tl.constexpr,
    stride_cos_b: tl.constexpr,
    stride_cos_h: tl.constexpr,
    stride_cos_t: tl.constexpr,
    stride_sin_b: tl.constexpr,
    stride_sin_h: tl.constexpr,
    stride_sin_t: tl.constexpr,
    stride_index_b: tl.constexpr,
    stride_k_b: tl.constexpr,
    stride_k_h: tl.constexpr,
    stride_k_s: tl.constexpr,
    stride_c_b: tl.constexpr,
    stride_c_h: tl.constexpr,
    stride_c_s: tl.constexpr,
    seq_len: tl.constexpr,
    num_heads: tl.constexpr,
    epsilon: tl.constexpr,
    cache_mode: tl.constexpr,
    block_size: tl.constexpr,
    k_d1: tl.constexpr,
    c_d1: tl.constexpr,
    cache_slots: tl.constexpr,
    has_k_scale: tl.constexpr,
    has_c_scale: tl.constexpr,
    has_k_offset: tl.constexpr,
    has_c_offset: tl.constexpr,
):
    # One program owns one (batch, head, token) row. This exposes the full
    # sequence parallelism needed by the long (up to 2048 token) cases.
    program = tl.program_id(0)
    tokens_per_batch = num_heads * seq_len
    batch = program // tokens_per_batch
    within_batch = program - batch * tokens_per_batch
    head = within_batch // seq_len
    token = within_batch - head * seq_len

    c_cols = tl.arange(0, 512)
    r_cols = tl.arange(0, 64)
    half_cols = r_cols % 32
    rope_real_col = 512 + half_cols * 2
    rope_imag_col = rope_real_col + 1
    rope_real_scale_col = half_cols * 2
    rope_imag_scale_col = rope_real_scale_col + 1

    gamma = tl.load(gamma_ptr + c_cols).to(tl.float32)
    if has_c_scale:
        c_scale = tl.load(c_scale_ptr + c_cols).to(tl.float32)
    if has_c_offset:
        c_offset = tl.load(c_offset_ptr + c_cols).to(tl.float32)

    kv_sequence = kv_ptr + batch * stride_kv_b + head * stride_kv_h
    cos_sequence = cos_ptr + batch * stride_cos_b + head * stride_cos_h
    sin_sequence = sin_ptr + batch * stride_sin_b + head * stride_sin_h
    index_sequence = index_ptr + batch * stride_index_b

    kv_row = kv_sequence + token * stride_kv_t

    ckv = tl.load(kv_row + c_cols).to(tl.float32)
    square_sum = tl.sum(ckv * ckv, axis=0)
    rms = tl.sqrt(square_sum * (1.0 / 512.0) + epsilon)
    ckv_out = ckv / rms * gamma
    if has_c_scale:
        ckv_out = ckv_out * c_scale
    if has_c_offset:
        ckv_out = ckv_out + c_offset

    real = tl.load(kv_row + rope_real_col).to(tl.float32)
    imag = tl.load(kv_row + rope_imag_col).to(tl.float32)
    if has_k_scale:
        real = real * tl.load(k_scale_ptr + rope_real_scale_col).to(tl.float32)
        imag = imag * tl.load(k_scale_ptr + rope_imag_scale_col).to(tl.float32)
    if has_k_offset:
        real = real + tl.load(k_offset_ptr + rope_real_scale_col).to(tl.float32)
        imag = imag + tl.load(k_offset_ptr + rope_imag_scale_col).to(tl.float32)

    cos_value = tl.load(cos_sequence + token * stride_cos_t + r_cols).to(tl.float32)
    sin_value = tl.load(sin_sequence + token * stride_sin_t + r_cols).to(tl.float32)
    rope_out = tl.where(
        r_cols < 32,
        real * cos_value - imag * sin_value,
        imag * cos_value + real * sin_value,
    )

    if cache_mode == 0:  # Norm: [B, H, cache_length, D]
        position = tl.load(index_sequence + token).to(tl.int32)
        valid = position >= 0
        k_dst = (
            k_cache_ptr
            + batch * stride_k_b
            + head * stride_k_h
            + position * stride_k_s
            + r_cols
        )
        c_dst = (
            ckv_cache_ptr
            + batch * stride_c_b
            + head * stride_c_h
            + position * stride_c_s
            + c_cols
        )
    elif cache_mode == 1:  # PA/PA_BNSD: global [position, head, D]
        position = tl.load(index_sequence + token).to(tl.int32)
        valid = (position >= 0) & (position < cache_slots)
        k_dst = k_cache_ptr + (position * num_heads + head) * 64 + r_cols
        c_dst = ckv_cache_ptr + (position * num_heads + head) * 512 + c_cols
    elif cache_mode == 2:  # PA_NZ: [block, H*D/16, block_size, 16]
        position = tl.load(index_sequence + token).to(tl.int32)
        valid = (position >= 0) & (position < cache_slots)
        page = position // block_size
        in_page = position - page * block_size
        k_dst = (
            k_cache_ptr
            + page * num_heads * k_d1 * block_size * 16
            + head * k_d1 * block_size * 16
            + (r_cols // 16) * block_size * 16
            + in_page * 16
            + r_cols % 16
        )
        c_dst = (
            ckv_cache_ptr
            + page * num_heads * c_d1 * block_size * 16
            + head * c_d1 * block_size * 16
            + (c_cols // 16) * block_size * 16
            + in_page * 16
            + c_cols % 16
        )
    elif cache_mode == 3:  # PA_BLK_BNSD: [block, block_size, head, D]
        page_entry = token // block_size
        page_position = tl.load(index_sequence + page_entry).to(tl.int32)
        page = page_position // block_size
        in_page = token - page_entry * block_size
        valid = (page_position >= 0) & (page < cache_slots)
        k_dst = (
            k_cache_ptr
            + (page * block_size * num_heads + in_page * num_heads + head) * 64
            + r_cols
        )
        c_dst = (
            ckv_cache_ptr
            + (page * block_size * num_heads + in_page * num_heads + head) * 512
            + c_cols
        )
    else:  # PA_BLK_NZ: [block, H*D/16, block_size, 16]
        page_entry = token // block_size
        page_position = tl.load(index_sequence + page_entry).to(tl.int32)
        page = page_position // block_size
        in_page = token - page_entry * block_size
        valid = (page_position >= 0) & (page < cache_slots)
        k_dst = (
            k_cache_ptr
            + page * num_heads * k_d1 * block_size * 16
            + head * k_d1 * block_size * 16
            + (r_cols // 16) * block_size * 16
            + in_page * 16
            + r_cols % 16
        )
        c_dst = (
            ckv_cache_ptr
            + page * num_heads * c_d1 * block_size * 16
            + head * c_d1 * block_size * 16
            + (c_cols // 16) * block_size * 16
            + in_page * 16
            + c_cols % 16
        )

    tl.store(k_dst, rope_out, mask=valid)
    tl.store(c_dst, ckv_out, mask=valid)


def kv_rmsnorm_rope_cache(
    kv: torch.Tensor,
    gamma: torch.Tensor,
    cos: torch.Tensor,
    sin: torch.Tensor,
    index: torch.Tensor,
    k_cache: torch.Tensor,
    ckv_cache: torch.Tensor,
    k_rope_scale: torch.Tensor = None,
    c_kv_scale: torch.Tensor = None,
    k_rope_offset: torch.Tensor = None,
    c_kv_offset: torch.Tensor = None,
    epsilon: float = 1e-5,
    cache_mode: str = "Norm",
    is_output_kv: bool = False,
) -> tuple[torch.Tensor, torch.Tensor]:
    mode = {
        "Norm": 0,
        "PA": 1,
        "PA_BNSD": 1,
        "PA_NZ": 2,
        "PA_BLK_BNSD": 3,
        "PA_BLK_NZ": 4,
    }[cache_mode]

    batch_size, num_heads, seq_len, _ = kv.shape
    # Paged modes expose the logical [block_num, block_size, H, D]
    # shape even when the backing memory is interpreted as NZ.
    block_size = k_cache.shape[1] if mode in (2, 3, 4) else 1
    if index.ndim > 1:
        index_batch_stride = index.stride(0)
    elif mode in (3, 4):
        index_batch_stride = triton.cdiv(seq_len, block_size)
    else:
        index_batch_stride = seq_len
    k_d1 = 4
    c_d1 = 32

    # Optional tensors are absent in all official cases. Supplying harmless
    # pointer placeholders keeps a single kernel signature for both variants.
    k_scale_arg = k_rope_scale if k_rope_scale is not None else gamma
    c_scale_arg = c_kv_scale if c_kv_scale is not None else gamma
    k_offset_arg = k_rope_offset if k_rope_offset is not None else gamma
    c_offset_arg = c_kv_offset if c_kv_offset is not None else gamma

    if mode == 0:
        block_t = 1
        while block_t < seq_len:
            block_t *= 2
        _kv_rmsnorm_rope_cache_norm_kernel[(seq_len, batch_size * num_heads)](
            kv,
            gamma,
            cos,
            sin,
            index,
            k_cache,
            ckv_cache,
            k_scale_arg,
            c_scale_arg,
            k_offset_arg,
            c_offset_arg,
            kv.stride(0),
            kv.stride(1),
            kv.stride(2),
            cos.stride(0),
            cos.stride(1),
            cos.stride(2) if cos.shape[2] != 1 else 0,
            sin.stride(0),
            sin.stride(1),
            sin.stride(2) if sin.shape[2] != 1 else 0,
            index_batch_stride,
            k_cache.stride(0),
            k_cache.stride(1),
            k_cache.stride(2),
            ckv_cache.stride(0),
            ckv_cache.stride(1),
            ckv_cache.stride(2),
            k_cache.shape[-2],
            seq_len,
            num_heads,
            epsilon,
            k_rope_scale is not None,
            c_kv_scale is not None,
            k_rope_offset is not None,
            c_kv_offset is not None,
            block_t=block_t,
        )
        return k_cache, ckv_cache

    cache_slots = k_cache.shape[0] if mode in (3, 4) else k_cache.shape[0] * k_cache.shape[1]

    _kv_rmsnorm_rope_cache_kernel[(batch_size * num_heads * seq_len,)](
        kv,
        gamma,
        cos,
        sin,
        index,
        k_cache,
        ckv_cache,
        k_scale_arg,
        c_scale_arg,
        k_offset_arg,
        c_offset_arg,
        kv.stride(0),
        kv.stride(1),
        kv.stride(2),
        cos.stride(0),
        cos.stride(1),
        cos.stride(2) if cos.shape[2] != 1 else 0,
        sin.stride(0),
        sin.stride(1),
        sin.stride(2) if sin.shape[2] != 1 else 0,
        index_batch_stride,
        k_cache.stride(0) if k_cache.ndim >= 4 else 0,
        k_cache.stride(1) if k_cache.ndim >= 4 else 0,
        k_cache.stride(2) if k_cache.ndim == 4 else 0,
        ckv_cache.stride(0) if ckv_cache.ndim >= 4 else 0,
        ckv_cache.stride(1) if ckv_cache.ndim >= 4 else 0,
        ckv_cache.stride(2) if ckv_cache.ndim == 4 else 0,
        seq_len,
        num_heads,
        epsilon,
        mode,
        block_size,
        k_d1,
        c_d1,
        cache_slots,
        k_rope_scale is not None,
        c_kv_scale is not None,
        k_rope_offset is not None,
        c_kv_offset is not None,
    )
    return k_cache, ckv_cache
