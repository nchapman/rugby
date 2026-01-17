// Simple single-threaded prime sieve (no async dependencies)
use std::io::{self, Write, BufWriter};

fn main() {
    let n: usize = std::env::args()
        .nth(1)
        .and_then(|s| s.parse().ok())
        .unwrap_or(100);

    let mut stdout = BufWriter::new(io::stdout());
    let mut primes: Vec<usize> = Vec::new();
    let mut candidate = 2;

    while primes.len() < n {
        let is_prime = primes.iter().all(|&p| candidate % p != 0);
        if is_prime {
            writeln!(stdout, "{}", candidate).unwrap();
            primes.push(candidate);
        }
        candidate += 1;
    }
}
