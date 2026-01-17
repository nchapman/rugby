fn fannkuch(n: usize) -> (i32, i32) {
    let mut perm = vec![0; n];
    let mut perm1: Vec<usize> = (0..n).collect();
    let mut count = vec![0; n];

    let mut max_flips = 0;
    let mut checksum: i32 = 0;
    let mut r = n;
    let mut sign = true;

    loop {
        // Generate next permutation
        while r != 1 {
            count[r - 1] = r;
            r -= 1;
        }

        // Copy and count flips
        perm.copy_from_slice(&perm1);
        let mut flips = 0;
        while perm[0] != 0 {
            let k = perm[0];
            // Reverse perm[0..k]
            perm[0..=k].reverse();
            flips += 1;
        }

        if flips > max_flips {
            max_flips = flips;
        }
        if sign {
            checksum += flips;
        } else {
            checksum -= flips;
        }

        // Next permutation
        loop {
            if r == n {
                return (checksum, max_flips);
            }
            let p0 = perm1[0];
            for i in 0..r {
                perm1[i] = perm1[i + 1];
            }
            perm1[r] = p0;

            count[r] -= 1;
            if count[r] > 0 {
                break;
            }
            r += 1;
        }
        sign = !sign;
    }
}

fn main() {
    let n: usize = std::env::args()
        .nth(1)
        .and_then(|s| s.parse().ok())
        .unwrap_or(7);

    let (checksum, max_flips) = fannkuch(n);
    println!("{}\nPfannkuchen({}) = {}", checksum, n, max_flips);
}
