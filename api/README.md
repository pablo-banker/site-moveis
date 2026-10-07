# Forma API

Backend da loja: catálogo, contas e pedidos persistidos em PostgreSQL, servido por Go 1.26.8, Fiber 3, Uber Fx e Zap. Swagger UI está em `/docs`, com contrato OpenAPI 3 em `/openapi.json`.

## Executar

Na raiz do projeto:

```sh
docker compose up -d --build
```

- API: `http://127.0.0.1:8081`
- Swagger: `http://127.0.0.1:8081/docs`
- PostgreSQL: `127.0.0.1:5433`, banco/usuário `forma`.
- A senha padrão `forma_local` é exclusiva do desenvolvimento. Pode ser substituída por `POSTGRES_PASSWORD` no ambiente do Compose.
- Portas publicadas apenas em loopback. Volume `forma_forma_postgres` preserva os dados entre execuções.

Para executar Go fora do container, suba apenas o banco e, em `/api`:

```sh
cp .env.example .env
set -a
. ./.env
set +a
go run ./cmd/server
```

O Go lê variáveis de ambiente; não carrega `.env` automaticamente. Não execute as duas formas da API ao mesmo tempo na porta 8081.

## Abstract Factory

A interface `application.RepositoryFactory` cria dois produtos abstratos relacionados: `ProductRepository` e `TaxonomyRepository`. A fábrica concreta `postgres.Factory` entrega a família inteira usando o mesmo pool. O serviço usa somente as abstrações. Nos testes, uma fábrica de memória substitui essa família sem alterar o serviço.

A referência é o [Abstract Factory do Refactoring Guru](https://refactoring.guru/pt-br/design-patterns/abstract-factory). O padrão organiza a criação dos repositórios; não implica criar classes diferentes para cada móvel cadastrado. Novos móveis são registros no banco.

Fluxo: **Fiber → serviço → interfaces → repositórios Postgres**. O Fx compõe a fábrica, gerencia o pool, abre a porta depois da conexão/migração e encerra servidor e banco em ordem inversa.

## Rotas

| Método | Rota | Comportamento |
| --- | --- | --- |
| GET | `/api/v1/products` | Lista paginada |
| GET | `/api/v1/products/:id` | Detalhe; 404 se inexistente/inativo |
| GET | `/api/v1/categories` | Categorias com slug e nome |
| GET | `/api/v1/rooms` | Ambientes com slug e nome |
| GET | `/health/live` | Processo ativo |
| GET | `/health/ready` | Conexão Postgres disponível |

Lista aceita `q`, `category` e `room` (slugs), `sort` (`relevance`, `price_asc`, `price_desc`), `limit` (1–100, padrão 24), `offset` (≥0). Busca literal, sem diferenciação de maiúsculas, por nome, material ou categoria. A resposta contém `data`, `total`, `limit`, `offset`. A contagem e a lista usam o mesmo snapshot de leitura. `priceCents` é inteiro; `price` mantém compatibilidade em reais com as telas existentes.

Consultas usam parâmetros SQL e ordenação permitida por lista fixa. Erros retornam código, mensagem pública e identificador de requisição. Zap registra método, caminho e duração, sem query, corpo ou credenciais.

## Migrações

SQLs embarcados no binário, aplicados em transação com lock consultivo e checksum em `schema_migrations`. `AUTO_MIGRATE=true` habilita aplicação no início; o padrão fora do Compose é `false`. Migrações aplicadas não devem ser editadas: crie um novo arquivo numerado. A segunda migração insere os oito móveis demonstrativos e seus acabamentos. Executá-la novamente não sobrescreve os dados.

Nenhuma migração destrutiva/down é executada automaticamente. Em produção, aplique migrações com uma etapa controlada da implantação e revise o seed demonstrativo antes de adotar dados comerciais.

## Validação

```sh
go test -race ./...
go vet ./...
TEST_DATABASE_URL='postgres://forma:forma_local@127.0.0.1:5433/forma?sslmode=disable' go test -race ./...
```

Os testes de integração criam um schema exclusivo temporário e o removem ao terminar. Sem `TEST_DATABASE_URL`, eles são explicitamente ignorados. Verificam migração idempotente, família de repositórios, filtros, ordenação, paginação, SQL parametrizado, 400/404, preços, health checks e alteração persistida aparecendo na API.

## Integração e limites desta etapa

SvelteKit busca o catálogo no servidor via `API_URL`, configurável em `/web/.env`. Por padrão usa `http://127.0.0.1:8081`. `CATALOG_SOURCE=demo` habilita explicitamente a prévia sem API; indisponibilidade do backend no modo API retorna 503, sem substituição silenciosa por fixtures.

As telas preservam busca e filtros locais sobre os produtos carregados, usando paginação da API para buscar o catálogo completo nesta etapa. Para um catálogo grande, o próximo ajuste será fazer filtros e paginação diretamente no servidor por página, evitando carregar tudo no layout. O pool de conexão e as credenciais ficam exclusivamente na API.

Sacola/favoritos continuam no navegador. Contas, pedidos e estoque por acabamento são persistidos. Os totais são calculados no servidor; frete usa cotação persistida e validada. Pagamento usa apenas um simulador local gratuito. Nenhuma cobrança ou entrega real é realizada. Swagger UI usa arquivos versionados do pacote público `swagger-ui-dist` via CDN; o contrato JSON local independe da CDN.


## Contas, pedidos e provedores

Outra Abstract Factory cria a família `IdentityRepository` + `OrderRepository`, usando o mesmo pool do Postgres. O serviço comercial depende dessas interfaces. `PaymentGateway` é injetado pelo Fx; a implementação `simulated.Gateway` não usa rede, cartões, chaves de API ou recursos bancários.

- `POST /api/v1/auth/register`: nome, e-mail e senha (mínimo 15 caracteres, máximo 128 bytes). Retorna 202 com mensagem genérica, inclusive para e-mail já cadastrado. Cria conta, desafio de confirmação e e-mail na mesma transação; não abre sessão automaticamente.
- `POST /api/v1/auth/login`: e-mail e senha.
- `GET /api/v1/auth/me`: cliente da sessão autenticada, incluindo `emailVerified`.
- `POST /api/v1/auth/verify/request` e `/confirm`: reenvio e confirmação de titularidade.
- `POST /api/v1/auth/reset/request` e `/confirm`: solicitação genérica e redefinição de senha.
- `GET /api/v1/auth/sessions`: sessões ativas, sem expor seus tokens.
- `POST /api/v1/auth/sessions/revoke`: encerra outras sessões, mantendo a atual.
- `POST /api/v1/auth/profile`: atualiza somente o nome da conta autenticada, com 2–100 caracteres.
- `POST /api/v1/auth/logout`: revoga a sessão atual no banco.
- `GET /api/v1/products/:id/variants`: estoque disponível por acabamento.
- `POST /api/v1/orders`: itens, endereço, shippingQuoteId e chave UUID de idempotência; preços e entrega são calculados no backend.
- `GET /api/v1/orders`: até 100 pedidos recentes da conta.
- `GET /api/v1/orders/:id`: apenas o proprietário pode consultar.
- `POST /api/v1/orders/:id/cancel`: cancela um pedido pendente e devolve estoque uma única vez.
- `POST /api/v1/orders/:id/payment`: `{}`; o provedor decide o resultado. Aprovação muda o pedido para `paid`. O adaptador simulado é configurado por `PAYMENT_SIMULATED_OUTCOME=approved|declined`. Campos como `outcome` enviados pelo cliente são rejeitados.

O frontend chama rotas do SvelteKit em `/backend`, que mantêm o token em cookie HttpOnly, SameSite=Lax; em produção usa __Host-forma-session com Secure obrigatório. Mutações verificam a origem. A API recebe Bearer no servidor; o token não é exposto em localStorage nem no JSON enviado ao navegador. Sessões expiram em sete dias e só o hash do token fica no banco.

Senhas usam Argon2id com salt aleatório de 16 bytes, 64 MiB, três iterações e paralelismo dois; comparação em tempo constante. O prefixo `argon2id$v1$` identifica os parâmetros fixos desta versão. Referências: [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) e [pacote Argon2 de Go](https://pkg.go.dev/golang.org/x/crypto/argon2). Login/cadastro limitam tentativas por conta no Postgres (8/minuto), além dos limites compartilhados por IP no Postgres. Argon2 aceita no máximo dois cálculos simultâneos por processo. JSON rejeita campos desconhecidos e corpos acima de 32 KiB. Os limites por IP e quotas de pedidos/frete são compartilhados no Postgres. Em produção, Caddy sobrescreve o IP encaminhado e o BFF assina IP, método, caminho e timestamp com HMAC. A API rejeita chamadas sem assinatura válida; não publique sua porta diretamente.

A reserva ocorre em transação, com bloqueio dos acabamentos em ordem consistente. Uma falha em qualquer item reverte a reserva inteira. Repetir a mesma chave e corpo retorna o mesmo pedido; reutilizar a chave com outros dados retorna conflito. Cancelamento e pagamento bloqueiam a linha do pedido para impedir transições concorrentes incompatíveis. Um pedido pago não pode ser cancelado por esta versão. Estoque inicial demonstrativo: cinco unidades por acabamento.

Pedidos pendentes reservam estoque por 30 minutos. A rotina executada a cada minuto cancela reservas vencidas e devolve o estoque em transação; bloqueios impedem devolução duplicada entre workers. Cadastro envia confirmação via outbox SMTP. Checkout exige e-mail confirmado. Recuperação usa token de uso único, expira em 30 minutos e revoga todas as sessões; confirmação expira em uma hora. Tokens ficam em fragmentos de URL, desafios guardam somente hashes e mensagens pendentes são criptografadas com AES-GCM. Lista de pedidos continua limitada aos 100 mais recentes.

## Testes comerciais

Além do catálogo, os testes isolados verificam cadastro, hashes, login incorreto, sessão revogada/expirada, autorização por proprietário, idempotência, totais em centavos, rollback de reserva, recusa/aprovação simulada e seis pedidos concorrentes disputando cinco unidades.


## Frete e mensagens

`POST /api/v1/shipping/quotes` exige autenticação e recebe `{ "cep": "88301401", "items": [{"id":"sofa-arco","finish":"Linho natural","quantity":1}] }`. A aplicação valida os itens, usa `ShippingProvider` injetado pelo Fx e salva a cotação, vinculada ao cliente e ao hash canônico de CEP + itens. A validade é limitada a 15 minutos. `POST /orders` recebe `shippingQuoteId`; a transação verifica proprietário, validade e fingerprint e persiste preço/prazo/provedor como snapshot no pedido. O cliente não envia valores de frete. Repetições idempotentes de pedidos já criados retornam o pedido mesmo após a cotação expirar.

O adaptador `simulated.Shipping` calcula valores em centavos por faixa de CEP e quantidade. As regras são ilustrativas e existem somente no backend. O frontend de checkout, produto e confirmação exibe preço/prazo da resposta, sem tabelas de frete locais. Para cotar no produto é necessário entrar na conta.

`POST /api/v1/contact` valida e persiste mensagens em `contact_messages`, com limitação de frequência. O frontend só confirma recebimento após sucesso da API. Notificação ao atendimento não está implementada.

Em `APP_ENV=production`, a API exige conexão Postgres com `sslmode=verify-full`, senha não padrão, CORS explícito em HTTPS e `AUTO_MIGRATE=false`. A documentação fica desabilitada por padrão (`ENABLE_API_DOCS=false`). A revisão e comandos de segurança estão em [security-review.md](../docs/security-review.md).

Veja [preparação para produção](../docs/production-readiness.md): a escolha de provedores não substitui as pendências operacionais identificadas.
