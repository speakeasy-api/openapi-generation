use std::io::{self, Read, Write};

fn main() {
    let mut input = String::new();
    io::stdin()
        .read_to_string(&mut input)
        .expect("failed to read stdin");

    match rubyfmt::format_buffer(&input) {
        Ok(formatted) => {
            io::stdout()
                .write_all(formatted.as_bytes())
                .expect("failed to write stdout");
        }
        Err(e) => {
            eprintln!("rubyfmt error: {:?}", e);
            std::process::exit(e.as_exit_code());
        }
    }
}
