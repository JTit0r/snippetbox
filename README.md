# Let's Go - Aplicativo Web Snippetbox em Go
[![CI](https://github.com/vancanhuit/snippetbox/actions/workflows/ci.yml/badge.svg)](https://github.com/vancanhuit/snippetbox/actions/workflows/ci.yml)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://github.com/pre-commit/pre-commit)

Projeto apresentado como avaliação final na UC Tópicos em Tecnologia da Computação IV (Aprendendo stacks de tecnologia: Ruby e Golang)

## Features
- Adicionadas tags aos snippets
- Implementada uma feature de pesquisa de snippets
- - Snippets podem ser pesquisados por título e tags
- - Os resultados são exibídos em uma nova página

## Local development
```bash
make db
make run
```

## Running tests
```bash
make testdb
make test
```

## Running with Docker Compose

```bash
docker compose up -d --build
```
