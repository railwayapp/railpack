import gleam/io
import gleam/list
import gleam/string
import simplifile

pub fn main() -> Nil {
  let assert Ok(files) = simplifile.read_directory("src")
  let sorted_files = list.sort(files, by: string.compare)
  list.each(sorted_files, io.println)
}
