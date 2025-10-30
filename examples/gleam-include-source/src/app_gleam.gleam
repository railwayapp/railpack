import gleam/io
import simplifile

pub fn main() -> Nil {
  let assert Ok(contents) = simplifile.read(from: "test.txt")
  io.println(contents)
}
