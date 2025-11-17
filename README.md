# PulseNote 
PulseNote é um TUI em Go (bubbletea) projetado para produtividade por teclado: editor rápido, hotkeys, busca full-text com SQLite FTS e arquitetura client-server modular.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) 
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/gustavo-silva98/PulseNote/go.yml)

## 📸 Demonstração

<div align=center style="width:500px; height:200px; overflow:hidden; border-radius:15px; margin:auto;">
  <img src="assets/Gif Completo ffmpeg.gif" alt="GIF com bordas arredondadas" style="border-radius:15px;" />
</div>

---

## ✨ Recursos
- Criar, listar e remover notas.
- Hotkeys para ativação de funcionalidades.
- Interface amigável no terminal.
- Armazenamento local com SQLite.
- Pesquisa de notas otimizada utilizando FTS.
- Server + Client ( dois executáveis ) - arquitetura leve para uso local. 

---

## 🚀 Instalação
### Pré-requisitos
- https://go.dev/dl/ **>= 1.25.4**
- Git instalado.

> **Observação:** releases com binários estão disponíveis na página de Releases — se prefere não compilar, baixe o `.zip`/`.exe` na release.

### Passos
```bash
# Clone o repositório em uma pasta
git clone https://github.com/gustavo-silva98/PulseNote
cd PulseNote

# O comando abaixo executa um script de compilação dos dois binários
go run ./install/install.go

# A pasta bin com os binários compilados será criado em uma pasta acima
cd ..

# Para utilização, execute o server.exe na pasta bin. Compilando ou via release.
./bin/server.exe
```

---
### 🔑 HotKeys 
- Ctrl + Shift + H -> Salvar Notas
- Ctrl + Shift + R -> Ler Notas
- Ctrl + Shift + K -> Finalizar Server
- Ctrl + Shift + D -> Busca avançada
---
### 📁 Localização do banco
- Por padrão, o arquivo do banco é localizado em data/banco.db, conforme estrutura.
 ```
├──PulseNote
│   ├──cmd
│	... Demais pastas
│   	
├──bin
│   	
├──data
    └──banco.db
```
---
## ⚠️ Compatibilidade (Wayland / Linux)
- Atualmente o PulseNote é **suportado somente no Windows Terminal**. Em ambientes com **Wayland** (por exemplo GNOME on Wayland) podem ocorrer problemas de captura de teclas e redimensionamento do terminal.
- É um desejo viabilizar a compatibilidade com Linux, porém ainda não foi possível a realização.
---
### 🛠 Tecnologias Utilizadas
- Go
- Bubbletea TUI Framework
- Lipgloss
- Hotkey
- SQLite

---

## 📦 Releases

Binários (client + server) são publicados nas Releases do GitHub. Cada release contém os executáveis e um ZIP com ambos os arquivos.
Confira: `https://github.com/gustavo-silva98/PulseNote/releases`

---

## 🧪 Testes & CI

Testes unitários cobrem o core do projeto. CI (GitHub Actions) roda os testes e build  — veja os workflows no diretório `.github/workflows`.

---
### 📄 Licença
Este projeto está sob a licença MIT.
