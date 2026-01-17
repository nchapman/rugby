use regex::bytes::Regex;
use std::io::{self, Read, Write, BufWriter};

fn main() -> io::Result<()> {
    let file_name = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "25000_in".to_string());

    let mut bytes = Vec::new();
    std::fs::File::open(&file_name)?.read_to_end(&mut bytes)?;
    let original_len = bytes.len();

    // Clean the input - remove headers and newlines
    let clean_re = Regex::new(r"(>[^\n]+)?\n").unwrap();
    let cleaned = clean_re.replace_all(&bytes, &b""[..]);
    let bytes = cleaned.to_vec();
    let cleaned_len = bytes.len();

    let mut stdout = BufWriter::new(io::stdout());

    // Variant patterns to count
    let variants = [
        "agggtaaa|tttaccct",
        "[cgt]gggtaaa|tttaccc[acg]",
        "a[act]ggtaaa|tttacc[agt]t",
        "ag[act]gtaaa|tttac[agt]ct",
        "agg[act]taaa|ttta[agt]cct",
        "aggg[acg]aaa|ttt[cgt]ccct",
        "agggt[cgt]aa|tt[acg]accct",
        "agggta[cgt]a|t[acg]taccct",
        "agggtaa[cgt]|[acg]ttaccct",
    ];

    for pattern in &variants {
        let re = Regex::new(pattern).unwrap();
        let count = re.find_iter(&bytes).count();
        writeln!(stdout, "{} {}", pattern, count)?;
    }

    // Substitutions
    let substitutions = [
        ("tHa[Nt]", "<4>"),
        ("aND|caN|Ha[DS]|WaS", "<3>"),
        ("a[NSt]|BY", "<2>"),
        ("<[^>]*>", "|"),
        (r"\|[^|][^|]*\|", "-"),
    ];

    let mut current = bytes;
    for (pattern, replacement) in &substitutions {
        let re = Regex::new(pattern).unwrap();
        current = re.replace_all(&current, replacement.as_bytes()).to_vec();
    }

    writeln!(stdout)?;
    writeln!(stdout, "{}", original_len)?;
    writeln!(stdout, "{}", cleaned_len)?;
    writeln!(stdout, "{}", current.len())?;

    Ok(())
}
