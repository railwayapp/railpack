use std::process::Command;

fn main() {
    let output = Command::new("rustc")
        .arg("--version")
        .output()
        .expect("Failed to execute rustc");

    let version = String::from_utf8(output.stdout)
        .expect("Invalid UTF-8 in rustc output")
        .trim()
        .to_string();

    let simplified_version = version
        .split('(')
        .next()
        .unwrap_or(&version)
        .trim();

    println!("cargo:rustc-env=RUST_VERSION={}", simplified_version);
}
