struct TreeNode {
    left: Option<Box<TreeNode>>,
    right: Option<Box<TreeNode>>,
}

impl TreeNode {
    fn check(&self) -> i32 {
        let mut ret = 1;
        if let Some(l) = &self.left {
            ret += l.check();
        }
        if let Some(r) = &self.right {
            ret += r.check();
        }
        ret
    }

    fn create(depth: i32) -> Box<TreeNode> {
        if depth > 0 {
            Box::new(TreeNode {
                left: Some(Self::create(depth - 1)),
                right: Some(Self::create(depth - 1)),
            })
        } else {
            Box::new(TreeNode {
                left: None,
                right: None,
            })
        }
    }
}

const MIN_DEPTH: i32 = 4;

fn main() {
    let n: i32 = std::env::args()
        .nth(1)
        .and_then(|s| s.parse().ok())
        .unwrap_or(10);

    let max_depth = if MIN_DEPTH + 2 > n { MIN_DEPTH + 2 } else { n };

    {
        let depth = max_depth + 1;
        let tree = TreeNode::create(max_depth + 1);
        println!("stretch tree of depth {}\t check: {}", depth, tree.check());
    }

    let long_lived_tree = TreeNode::create(max_depth);

    let mut d = MIN_DEPTH;
    while d <= max_depth {
        let iterations = 1 << ((max_depth - d + MIN_DEPTH) as u32);
        let mut chk = 0;
        for _ in 0..iterations {
            let a = TreeNode::create(d);
            chk += a.check();
        }
        println!("{}\t trees of depth {}\t check: {}", iterations, d, chk);
        d += 2;
    }

    println!(
        "long lived tree of depth {}\t check: {}",
        max_depth,
        long_lived_tree.check()
    );
}
