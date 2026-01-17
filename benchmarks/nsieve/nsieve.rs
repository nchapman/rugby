fn nsieve(n: usize) {
    let mut flags = vec![false; n];
    let mut count = 0;
    for i in 2..n {
        if !flags[i] {
            count += 1;
            let mut j = i << 1;
            while j < n {
                flags[j] = true;
                j += i;
            }
        }
    }
    println!("Primes up to {:8} {:8}", n, count);
}

fn main() {
    let n: usize = std::env::args()
        .nth(1)
        .and_then(|s| s.parse().ok())
        .unwrap_or(4);

    for i in 0..3 {
        nsieve(10000 << (n - i));
    }
}
