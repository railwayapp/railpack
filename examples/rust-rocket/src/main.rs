#[macro_use]
extern crate rocket;

#[get("/")]
fn index() -> String {
    format!("Hello, from Rocket! ({})", env!("RUST_VERSION"))
}

#[launch]
fn rocket() -> _ {
    rocket::build().mount("/", routes![index])
}
