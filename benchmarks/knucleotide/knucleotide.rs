use std::collections::HashMap;
use std::io::{BufRead, BufReader};

fn main() {
    let dna = read_input();

    print!("{}", write_frequencies(&dna, 1));
    print!("{}", write_frequencies(&dna, 2));
    println!("{}", write_count(&dna, "GGT"));
    println!("{}", write_count(&dna, "GGTA"));
    println!("{}", write_count(&dna, "GGTATT"));
    println!("{}", write_count(&dna, "GGTATTTTAATT"));
    println!("{}", write_count(&dna, "GGTATTTTAATTTATAGT"));
}

fn read_input() -> String {
    let file_name = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "25000_in".to_string());

    let file = std::fs::File::open(file_name).unwrap();
    let reader = BufReader::new(file);
    let mut data = String::new();
    let mut three_reached = false;

    for line in reader.lines() {
        let line = line.unwrap();
        if three_reached {
            if !line.starts_with('>') {
                data.push_str(&line.to_uppercase());
            }
        } else if line.starts_with(">THREE") {
            three_reached = true;
        }
    }

    data
}

fn frequency(seq: &str, length: usize) -> HashMap<&str, usize> {
    let mut counts = HashMap::new();
    for i in 0..=seq.len().saturating_sub(length) {
        if i + length <= seq.len() {
            *counts.entry(&seq[i..i + length]).or_insert(0) += 1;
        }
    }
    counts
}

fn write_frequencies(seq: &str, length: usize) -> String {
    let counts = frequency(seq, length);
    let total = seq.len() - length + 1;

    let mut sorted: Vec<_> = counts.iter().collect();
    sorted.sort_by(|a, b| {
        b.1.cmp(a.1).then_with(|| a.0.cmp(b.0))
    });

    let mut result = String::new();
    for (key, value) in sorted {
        result.push_str(&format!("{} {:.3}\n", key, 100.0 * *value as f64 / total as f64));
    }
    result.push('\n');
    result
}

fn write_count(seq: &str, nucleotide: &str) -> String {
    let counts = frequency(seq, nucleotide.len());
    let count = counts.get(nucleotide).copied().unwrap_or(0);
    format!("{}\t{}", count, nucleotide)
}
