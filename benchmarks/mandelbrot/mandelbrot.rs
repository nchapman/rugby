use md5::{Md5, Digest};

const MAX_ITER: usize = 50;
const VLEN: usize = 8;

fn mbrot8(cr: &[f64; 8], civ: f64) -> u8 {
    let ci = [civ; 8];
    let mut zr = [0.0f64; 8];
    let mut zi = [0.0f64; 8];
    let mut tr = [0.0f64; 8];
    let mut ti = [0.0f64; 8];
    let mut absz = [0.0f64; 8];

    for _ in 0..MAX_ITER / 5 {
        for _ in 0..5 {
            for k in 0..8 {
                zi[k] = (zr[k] + zr[k]) * zi[k] + ci[k];
                zr[k] = tr[k] - ti[k] + cr[k];
                tr[k] = zr[k] * zr[k];
                ti[k] = zi[k] * zi[k];
            }
        }

        for k in 0..8 {
            absz[k] = tr[k] + ti[k];
        }

        if absz.iter().all(|&t| t > 4.0) {
            return 0;
        }
    }

    let mut accu = 0u8;
    for (i, &t) in absz.iter().enumerate() {
        if t <= 4.0 {
            accu |= 0x80 >> i;
        }
    }
    accu
}

fn main() {
    let size: usize = std::env::args()
        .nth(1)
        .and_then(|s| s.parse().ok())
        .unwrap_or(200);

    // Round size to multiple of 8
    let size = (size + VLEN - 1) / VLEN * VLEN;
    let chunk_size = size / VLEN;
    let inv = 2.0 / size as f64;

    let mut xloc = vec![[0.0f64; 8]; chunk_size];
    for i in 0..size {
        xloc[i / VLEN][i % VLEN] = i as f64 * inv - 1.5;
    }

    println!("P4\n{} {}", size, size);

    let mut pixels = vec![0u8; size * chunk_size];
    for chunk_id in 0..size {
        let ci = chunk_id as f64 * inv - 1.0;
        let offset = chunk_id * chunk_size;
        for i in 0..chunk_size {
            let r = mbrot8(&xloc[i], ci);
            if r > 0 {
                pixels[offset + i] = r;
            }
        }
    }

    let mut hasher = Md5::new();
    hasher.update(&pixels);
    let result = hasher.finalize();
    println!("{:x}", result);
}
