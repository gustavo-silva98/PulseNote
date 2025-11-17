# PulseNote 
Um aplicativo simples de interface de texto (TUI) para gerenciar notas diretamente no terminal, desenvolvido em Go.

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) 

## 📸 Demonstração

<div align=center style="width:500px; height:200px; overflow:hidden; border-radius:15px; margin:auto;">
  <img src="assets/Gif Completo ffmpeg.gif" alt="GIF com bordas arredondadas" style="border-radius:15px;" />
</div>

---

## ✨ Recursos
- Criar, listar e remover notas.
- Hotkeys para ativação de funcionalidades.
- Interface amigável no terminal.
- Armazenamento local simples (SQLite).
- Pesquisa de notas otimizada utilizando FTS.

---

## 🚀 Instalação
### Pré-requisitos
- https://go.dev/dl/ **>= 1.25.4**
- Git instalado.

### Passos
```bash
git clone https://github.com/gustavo-silva98/PulseNote
cd PulseNote
go run ./install/install.go
cd ..
./bin/server.exe
```
---
### 🎇 HotKeys 
- Ctrl + Shift + H -> Salvar Notas
- Ctrl + Shift + R -> Ler Notas
- Ctrl + Shift + K -> Finalizar Server
- Ctrl + Shift + D -> Busca avançada

---
### 🛠 Tecnologias Utilizadas
- Go
- Bubbletea TUI Framework
- Lipgloss
- Hotkey
- SQLite

### 📄 Licença
Este projeto está sob a licença MIT.
