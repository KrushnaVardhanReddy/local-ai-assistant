mod windows;
mod macos;
mod linux;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    tracing::info!("Native Agent initialized");
    println!("Native Agent initialized");
}
