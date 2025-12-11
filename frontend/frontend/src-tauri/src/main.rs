// src-tauri/src/main.rs
fn main() {
    tauri::Builder::default()
        .setup(|_app| {
            std::thread::spawn(|| {
                // Launch Go from project root (D:\pramodn\Voidspark-main\)
                let status = std::process::Command::new("go")
                    .current_dir("../../..")  // <-- critical
                    .args(["run", "cmd/voidspark/main.go"])
                    .status();

                match status {
                    Ok(s) if s.success() => println!("✅ Go backend ready"),
                    Ok(s) => eprintln!("❌ Go failed: code {}", s.code().unwrap_or(-1)),
                    Err(e) => eprintln!("❌ Go launch error: {}", e),
                }
            });
            println!("✅ Tauri UI ready");
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("Tauri init failed");
}