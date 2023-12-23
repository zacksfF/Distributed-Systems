use std::fs;
use std::path::{Path, PathBuf};
use std::sync::{Arc, Mutex};
use std::thread;

fn file_search(root: &Path, filename: &str, matches: Arc<Mutex<Vec<PathBuf>>>) {
    println!("Searching in {:?}", root);
    if let Ok(entries) = fs::read_dir(root) {
        for entry in entries {
            if let Ok(entry) = entry {
                let path = entry.path();
                if let Some(file_name) = path.file_name() {
                    if file_name.to_string_lossy().contains(filename) {
                        let mut matches = matches.lock().unwrap();
                        matches.push(path.clone());
                    }
                }
                if path.is_dir() {
                    let matches_clone = Arc::clone(&matches);
                    let root_clone = root.to_path_buf();
                    let filename_clone = filename.to_string(); // Clone filename
                    thread::spawn(move || {
                        file_search(&root_clone, &filename_clone, matches_clone); // Pass cloned filename
                    });
                }
            }
        }
    }
}

fn main() {
    let root_path = Path::new("/Users/zakariasaif");
    let filename_to_search = "Go";
    let matches = Arc::new(Mutex::new(Vec::new()));

    file_search(root_path, filename_to_search, Arc::clone(&matches));

    // Wait for all threads to finish
    thread::sleep(std::time::Duration::from_secs(1));

    // Print matched files
    let matches = matches.lock().unwrap();
    for file in matches.iter() {
        println!("Matched: {:?}", file);
    }
}
