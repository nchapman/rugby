use std::collections::HashMap;

const A: u32 = 1103515245;
const C: u32 = 12345;
const M: u32 = 1 << 31;

struct LCG {
    seed: u32,
}

impl LCG {
    fn new(seed: u32) -> Self {
        LCG { seed }
    }

    fn next(&mut self) -> u32 {
        let (v, _) = A.overflowing_mul(self.seed);
        let (v, _) = v.overflowing_add(C);
        self.seed = v % M;
        self.seed
    }
}

struct LRUNode {
    key: u32,
    value: u32,
    prev: Option<usize>,
    next: Option<usize>,
}

struct LRU {
    size: usize,
    cache: HashMap<u32, usize>,
    nodes: Vec<LRUNode>,
    head: Option<usize>,
    tail: Option<usize>,
    free: Vec<usize>,
}

impl LRU {
    fn new(size: usize) -> Self {
        LRU {
            size,
            cache: HashMap::with_capacity(size),
            nodes: Vec::with_capacity(size),
            head: None,
            tail: None,
            free: Vec::new(),
        }
    }

    fn get(&mut self, key: u32) -> Option<u32> {
        if let Some(&idx) = self.cache.get(&key) {
            let value = self.nodes[idx].value;
            self.move_to_end(idx);
            Some(value)
        } else {
            None
        }
    }

    fn put(&mut self, key: u32, value: u32) {
        if let Some(&idx) = self.cache.get(&key) {
            self.nodes[idx].value = value;
            self.move_to_end(idx);
            return;
        }

        if self.cache.len() == self.size {
            self.evict();
        }

        let idx = if let Some(free_idx) = self.free.pop() {
            self.nodes[free_idx] = LRUNode {
                key,
                value,
                prev: None,
                next: None,
            };
            free_idx
        } else {
            let idx = self.nodes.len();
            self.nodes.push(LRUNode {
                key,
                value,
                prev: None,
                next: None,
            });
            idx
        };

        self.cache.insert(key, idx);
        self.add_to_end(idx);
    }

    fn move_to_end(&mut self, idx: usize) {
        self.remove_node(idx);
        self.add_to_end(idx);
    }

    fn remove_node(&mut self, idx: usize) {
        let prev = self.nodes[idx].prev;
        let next = self.nodes[idx].next;

        if let Some(prev_idx) = prev {
            self.nodes[prev_idx].next = next;
        } else {
            self.head = next;
        }

        if let Some(next_idx) = next {
            self.nodes[next_idx].prev = prev;
        } else {
            self.tail = prev;
        }
    }

    fn add_to_end(&mut self, idx: usize) {
        self.nodes[idx].prev = self.tail;
        self.nodes[idx].next = None;

        if let Some(tail_idx) = self.tail {
            self.nodes[tail_idx].next = Some(idx);
        }
        self.tail = Some(idx);

        if self.head.is_none() {
            self.head = Some(idx);
        }
    }

    fn evict(&mut self) {
        if let Some(head_idx) = self.head {
            let key = self.nodes[head_idx].key;
            self.cache.remove(&key);
            self.remove_node(head_idx);
            self.free.push(head_idx);
        }
    }
}

fn main() {
    let size: usize = std::env::args()
        .nth(1)
        .and_then(|s| s.parse().ok())
        .unwrap_or(100);

    let n: usize = std::env::args()
        .nth(2)
        .and_then(|s| s.parse().ok())
        .unwrap_or(10000);

    let modular = size as u32 * 10;
    let mut rng0 = LCG::new(0);
    let mut rng1 = LCG::new(1);
    let mut lru = LRU::new(size);
    let mut hit = 0;
    let mut missed = 0;

    for _ in 0..n {
        let n0 = rng0.next() % modular;
        lru.put(n0, n0);

        let n1 = rng1.next() % modular;
        if lru.get(n1).is_some() {
            hit += 1;
        } else {
            missed += 1;
        }
    }

    println!("{}", hit);
    println!("{}", missed);
}
