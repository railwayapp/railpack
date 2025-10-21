use openssl as _;
use ring as _;

fn main() {
    println!("hello from rust system deps ({})", env!("RUST_VERSION"))
}
