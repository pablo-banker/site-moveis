 _____  ___  ____  __  __    _
|  ___|/ _ \|  _ \|  \/  |  / \
| |_  | | | | |_) | |\/| | / _ \
|  _| | |_| |  _ <| |  | |/ ___ \
|_|    \___/|_| \_\_|  |_/_/   \_\

          Uma nova forma de habitar.

Uma loja full-stack com identidade minimalista, cantos arredondados e uma experiência editorial: ambientes, histórias das peças, animações por scroll e parallax em camadas.

**Contas, catálogo, estoque e pedidos são persistidos no Postgres.** Somente os provedores de pagamento e envio simulam o processamento no servidor; o frontend mantém um fluxo normal de compra.

[Executar](#executar-localmente) · [Stack](#stack) · [Estrutura](#estrutura) · [Documentação](#documentação) · [Publicação](#publicação)

## O que você encontra

- **Explorar:** busca e filtros de móveis, acabamentos, detalhes e ambientes.
- **Comprar:** sacola, cotação de frete no servidor, reserva de estoque e acompanhamento dos pedidos.
- **Sua conta:** cadastro, confirmação de e-mail, recuperação de senha, edição do nome e gestão de sessões.
- **Sua seleção:** favoritos e sacola salvos neste navegador.

## Stack

| Camada | Tecnologias |
| --- | --- |
| Frontend `/web` | SvelteKit 2 · Svelte 5 · Tailwind CSS 4 · Bits UI · Lenis |
| Backend `/api` | Go · Fiber · Uber Fx · Zap · Swagger / OpenAPI |
| Dados | PostgreSQL 17 · migrações SQL versionadas |
| Infraestrutura | Docker Compose · Caddy / HTTPS · SMTP · Prometheus · backups com age |

A arquitetura usa **Abstract Factory** para compor famílias de repositórios, seguindo a [referência do Refactoring Guru](https://refactoring.guru/pt-br/design-patterns/abstract-factory). Os adaptadores ficam separados das regras de negócio.

## Executar localmente

Você precisa de **Docker com Compose** e **Node.js 24**. Execute os comandos na raiz do projeto.

**1. Suba a API, o banco e o serviço local de e-mails:**

```sh
docker compose up -d --build
```

**2. Instale as dependências e inicie o frontend:**

```sh
npm --prefix web ci
npm --prefix web run dev
```

| Serviço | Endereço |
| --- | --- |
| Loja | [127.0.0.1:5173](http://127.0.0.1:5173) |
| API | [127.0.0.1:8081](http://127.0.0.1:8081) |
| Swagger | [127.0.0.1:8081/docs](http://127.0.0.1:8081/docs) |
| E-mails / Mailpit | [127.0.0.1:8025](http://127.0.0.1:8025) |
| Postgres | `127.0.0.1:5433` |

Crie sua conta na loja e abra o e-mail de confirmação no **Mailpit**. No desenvolvimento, as mensagens ficam nessa caixa local, sem envio externo. O volume do Postgres preserva os dados entre execuções.

As configurações opcionais estão em [web/.env.example](web/.env.example) e [api/.env.example](api/.env.example). Para executar Go fora do Docker, consulte o [guia da API](api/README.md).

## Estrutura

```text
.
├── api/                       # API Go, regras de negócio e adaptadores
│   ├── internal/              # Catálogo, contas, pedidos e infraestrutura
│   ├── migrations/            # Schema e dados iniciais
│   └── docs/                  # Contrato OpenAPI
├── web/                       # Interface SvelteKit e BFF
│   ├── src/lib/               # Componentes, estado e tokens globais
│   └── static/background/     # Camadas do parallax
├── deploy/                    # HTTPS, SMTP, métricas e backup
├── docs/                      # Design, arquitetura e guias
└── compose.production.yaml    # Configuração separada de produção
```

## Verificar

Frontend:

```sh
npm --prefix web run check
npm --prefix web run build
```

Backend, com Go 1.26.8 e o banco local ativo:

```sh
cd api
TEST_DATABASE_URL='postgres://forma:forma_local@127.0.0.1:5433/forma?sslmode=disable' go test -race ./...
go vet ./...
```

Os testes de integração usam schemas temporários isolados. Se você mudou a senha local do banco, ajuste a URL. O [pipeline](.github/workflows/security.yml) também executa auditorias de dependências e scans das imagens.

## Documentação

| Quero entender… | Guia |
| --- | --- |
| API, rotas e migrações | [Backend](api/README.md) |
| Organização e padrões | [Arquitetura](docs/architecture.md) |
| Paleta, tipografia e componentes | [Design system](docs/design-system.md) |
| Animações e camadas do fundo | [Movimento](docs/motion.md) · [Parallax](docs/parallax.md) |
| Controles e validações de segurança | [Revisão de segurança](docs/security-review.md) |
| Deploy, e-mails, backup e monitoramento | [Guia de publicação](docs/publication.md) |
| Limites atuais e requisitos operacionais | [Preparação para produção](docs/production-readiness.md) |

## Publicação

A configuração inclui HTTPS, API privada, credenciais separadas, SMTP com TLS, métricas e backup criptografado. **Hospedagem, domínio e serviço SMTP ainda precisam ser definidos**, assim como a ativação de backups e alertas no destino.

Siga o [guia de publicação](docs/publication.md) e confira os [limites operacionais](docs/production-readiness.md). Os produtos, preços e fotos iniciais são referências e devem ser revisados antes de abrir acesso público. Pagamentos e entregas reais permanecem fora desta configuração.
