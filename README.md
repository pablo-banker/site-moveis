# Forma — Site de móveis

Loja de móveis com frontend em **SvelteKit 2, Svelte 5, Tailwind CSS 4 e Bits UI** e backend em **Go, Postgres, Uber Fx, Fiber, Zap, Swagger e Docker**. Identidade minimalista: branco, grafite e cinza. A loja utiliza provedores de pagamento e envio simulados no servidor.

## Executar

```sh
docker compose up -d --build
cd web
npm install
npm run dev
```

Site: `http://127.0.0.1:5173`. API: `http://127.0.0.1:8081`. [Swagger local](http://127.0.0.1:8081/docs). Postgres local usa a porta 5433 para evitar conflito com outros projetos. E-mails locais ficam no [Mailpit](http://127.0.0.1:8025), sem envio externo.

O catálogo agora vem do Postgres, carregado no servidor do SvelteKit. Configuração e execução alternativa: [API](api/README.md). Para revisar só a interface, configure `CATALOG_SOURCE=demo` em `web/.env`.

## Etapa atual

- Catálogo persistido, categorias, ambientes e detalhe dos produtos.
- Abstract Factory para a família de repositórios, composta pelo Fx.
- Migrações versionadas, health checks, erros padronizados e logs estruturados.
- Layout editorial, transições de página, revelações por scroll e parallax em camadas PNG.
- Sacola e favoritos locais; contas, pedidos e estoque persistidos, com checkout e pagamento via provedor.

Produtos, fotos, preços e condições são ilustrativos. Contas, sessões e pedidos são reais e persistidos. Nenhum pagamento ou entrega real é processado. Imagens de referência externas e Swagger UI via CDN precisam de conexão.

## Validar

```sh
npm run check --prefix web
npm run build --prefix web
cd api
go test -race ./...
go vet ./...
```

Consulte os comandos para testes com Postgres em [api/README.md](api/README.md).

## Organização

- `api/internal/catalog`: domínio, serviço, fábrica e adaptadores HTTP/Postgres.
- `api/internal/platform`: configuração, conexão, logs e ciclo de vida.
- `api/migrations`: schema e catálogo inicial versionados.
- `api/docs`: contrato OpenAPI e Swagger UI.
- `web/src/lib/styles/tokens.css`: paleta e demais tokens globais.
- `web/src/lib/server/catalog.ts`: cliente do catálogo no servidor.
- `web/src/lib/catalog.ts`: tipos, utilitários e fixtures do modo demo.
- `web/src/lib/shop.ts`: sacola/favoritos no navegador.
- `web/static/background`: camadas do parallax.

Veja [design](docs/design-system.md), [parallax](docs/parallax.md), [movimento](docs/motion.md) e [arquitetura](docs/architecture.md).


Frete cotado pela API, persistido com validade, validado ao criar o pedido. O frontend utiliza um fluxo normal de compra e não escolhe resultados de pagamento. Consulte [o estado de preparação para produção](docs/production-readiness.md) antes de publicar.

## Publicação

Configuração Docker com Caddy/HTTPS, adapter Node, SMTP, credenciais separadas, limites compartilhados, métricas e backup criptografado: [guia de publicação](docs/publication.md). Hospedagem, domínio e SMTP ainda precisam ser escolhidos. Somente os provedores de pagamento/envio permanecem simulados no servidor.
