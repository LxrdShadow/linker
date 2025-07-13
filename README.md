# Linker

A fast and efficient terminal-based file transfer utility written in **Go**. Supports sending and receiving files or entire directories over a **TCP** connection with chunked transfer, progress reporting, and error handling.

## Table of Contents

- [Features](#features)
- [Screenshots](#screenshots)
- [Installation](#installation)
- [Usage](#usage)
- [Planned Features](#planned-features)
- [Contributing](#contributing)
- [License](#license)
- [Author](#author)

---

## Features
- ✔️ Send and receive **single files** or **entire directories**
- ✔️ Supports **binary** and **text files**
- ✔️ Uses **chunked transfer** for large files
- ✔️ **Progress reporting** for file transfers
- ✔️ Doesn't use external libraries

---

## Screenshots

**Sender**
![Sender Screenshot](https://github.com/user-attachments/assets/cd64f4d4-7a54-4d93-929b-6033f0c8e5b2)
**Receiver**
![Receiver screenshot](https://github.com/user-attachments/assets/91240eb2-4a7c-4e8a-944c-a611172c8bb3)

---

## Installation
### **Prerequisites**
- **Go** (>= 1.22) installed on your system
- **make**
- A LAN connection for file transfer

### **Build from Source**
```sh
git clone https://github.com/LxrdShadow/linker.git
cd linker
make build
```
This generates a single binary **lnkr**.

---

## Usage
### **Start the Server (sender)**
```sh
./lnkr send -port 9090 example.txt
```
This will launch the server and display a message **Listening on [your-ip-address]**

### **Receive the files**
```sh
./lnkr receive -addr [ip-of-server]
```
By default, it saves received files in the current directory.

## Planned Features

- ✅ Multi-file support
- ✅ Directory transfer support
- ⏳ A proper installation process
- ⏳ Compression before sending
- ⏳ Secure transfer (TLS encryption)
- ⏳ Authentication (password-protected transfers)
- ⏳ Terminal User Interface (TUI)
- ⏳ Local web interface

---

## Contributing
Pull requests are welcome! Open an issue if you find a bug or have suggestions.

---

## Licence

Linker is licensed under the [MIT License](LICENSE). You are free to use, modify, and distribute this software, profided proper attribution is given.

---

## Author

👤 Idealy Andritiana
GitHub: [LxrdShadow](https://github.com/LxrdShadow)
Email: andritiana.idealy@gmail.com

---

Show your support by giving a star ⭐ to this repository and giving some feedbacks.
