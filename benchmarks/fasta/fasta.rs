use std::io::{self, BufWriter, Write};

const IM: i64 = 139968;
const IA: i64 = 3877;
const IC: i64 = 29573;
const LINE_WIDTH: usize = 60;

const ALU: &[u8] = b"GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAAAA";

struct Random {
    last: i64,
}

impl Random {
    fn new(seed: i64) -> Self {
        Random { last: seed }
    }

    fn next(&mut self) -> f64 {
        self.last = (self.last * IA + IC) % IM;
        self.last as f64 / IM as f64
    }
}

struct AminoAcid {
    c: u8,
    p: f64,
}

fn repeat_fasta<W: Write>(out: &mut W, seq: &[u8], mut n: usize) -> io::Result<()> {
    let length = seq.len();
    let mut pos = 0;

    while n > 0 {
        let line_len = n.min(LINE_WIDTH);

        for i in 0..line_len {
            out.write_all(&[seq[(pos + i) % length]])?;
        }
        out.write_all(b"\n")?;

        pos = (pos + line_len) % length;
        n -= line_len;
    }
    Ok(())
}

fn random_fasta<W: Write>(
    out: &mut W,
    genelist: &[AminoAcid],
    mut n: usize,
    rng: &mut Random,
) -> io::Result<()> {
    // Build cumulative probabilities
    let mut cum = 0.0;
    let probs: Vec<f64> = genelist
        .iter()
        .map(|aa| {
            cum += aa.p;
            cum
        })
        .collect();

    while n > 0 {
        let line_len = n.min(LINE_WIDTH);

        for _ in 0..line_len {
            let r = rng.next();
            for (k, prob) in probs.iter().enumerate() {
                if *prob >= r {
                    out.write_all(&[genelist[k].c])?;
                    break;
                }
            }
        }
        out.write_all(b"\n")?;

        n -= line_len;
    }
    Ok(())
}

fn main() -> io::Result<()> {
    let n: usize = std::env::args()
        .nth(1)
        .and_then(|s| s.parse().ok())
        .unwrap_or(1000);

    let stdout = io::stdout();
    let mut out = BufWriter::new(stdout.lock());

    let iub = vec![
        AminoAcid { c: b'a', p: 0.27 },
        AminoAcid { c: b'c', p: 0.12 },
        AminoAcid { c: b'g', p: 0.12 },
        AminoAcid { c: b't', p: 0.27 },
        AminoAcid { c: b'B', p: 0.02 },
        AminoAcid { c: b'D', p: 0.02 },
        AminoAcid { c: b'H', p: 0.02 },
        AminoAcid { c: b'K', p: 0.02 },
        AminoAcid { c: b'M', p: 0.02 },
        AminoAcid { c: b'N', p: 0.02 },
        AminoAcid { c: b'R', p: 0.02 },
        AminoAcid { c: b'S', p: 0.02 },
        AminoAcid { c: b'V', p: 0.02 },
        AminoAcid { c: b'W', p: 0.02 },
        AminoAcid { c: b'Y', p: 0.02 },
    ];

    let homosapiens = vec![
        AminoAcid { c: b'a', p: 0.3029549426680 },
        AminoAcid { c: b'c', p: 0.1979883004921 },
        AminoAcid { c: b'g', p: 0.1975473066391 },
        AminoAcid { c: b't', p: 0.3015094502008 },
    ];

    let mut rng = Random::new(42);

    out.write_all(b">ONE Homo sapiens alu\n")?;
    repeat_fasta(&mut out, ALU, 2 * n)?;

    out.write_all(b">TWO IUB ambiguity codes\n")?;
    random_fasta(&mut out, &iub, 3 * n, &mut rng)?;

    out.write_all(b">THREE Homo sapiens frequency\n")?;
    random_fasta(&mut out, &homosapiens, 5 * n, &mut rng)?;

    Ok(())
}
