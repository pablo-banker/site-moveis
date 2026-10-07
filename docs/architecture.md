# Arquitetura

Manter `/api` e `/web` independentes, com contratos HTTP documentados via Swagger. PostgreSQL é a fonte persistente de dados; credenciais e acesso ao banco ficam no backend.

## API

```text
api/
  cmd/server/             entrada da aplicação e composição Fx
  internal/
    platform/            configuração, PostgreSQL, Zap e servidor Fiber
    catalog/             produtos, categorias, variantes e ambientes
      domain/            entidades e regras independentes de HTTP e SQL
      application/       casos de uso e interfaces consumidas por eles
      adapters/http/     handlers Fiber e DTOs
      adapters/postgres/ persistência
    identity/            autenticação e autorização, quando definidas
  migrations/            evolução versionada do banco
  docs/                  contrato OpenAPI e Swagger UI
  Dockerfile
```

O módulo de pedidos atenderá ao modelo de loja online escolhido. Começar pelo catálogo; criar interfaces nos pontos de troca reais, sem uma abstração para cada função.

Fluxo: handler Fiber → caso de uso → interface de repositório → implementação PostgreSQL. Fx compõe as dependências e gerencia início e encerramento. Zap recebe logs estruturados com identificador de requisição, sem dados sensíveis. Swagger documenta os contratos públicos.

## Padrões de projeto

Referência: [catálogo de padrões em Go do Refactoring Guru](https://refactoring.guru/pt-br/design-patterns/go). O catálogo orienta decisões de implementação; a divisão de pastas acima é uma proposta arquitetural do projeto.

| Padrão | Aplicação quando necessária |
| --- | --- |
| Abstract Factory | Criar a família compatível de repositórios; implementado no catálogo |
| Adapter | Encapsular provedores de pagamento, frete, arquivos ou mensagens |
| Strategy | Alternar cálculo de frete ou descontos por uma interface pequena |
| State | Controlar transições de pedidos se o fluxo justificar essa abstração |
| Decorator | Adicionar métricas ou cache a uma implementação sem alterar o caso de uso |

Aplicar apenas quando houver necessidade concreta. Construtores `New...` injetados pelo Fx atendem à composição inicial; não são automaticamente o padrão Factory Method. Evitar Singleton global para banco e logger: usar o escopo e o ciclo de vida do Fx. Repositórios não são um padrão listado nesse catálogo e são uma decisão de separação da persistência.

## Web

```text
web/src/
  routes/                páginas e layout raiz
  lib/
    components/ui/       componentes com tokens e primitivas Bits UI
    components/catalog/  componentes específicos de móveis
    styles/tokens.css    cores, tipografia, espaços e medidas globais
    api/                 cliente HTTP e tipos do contrato
  app.css                entrada global de estilos
```

Priorizar renderização no servidor nas páginas públicas do catálogo. Os preços, estoque, autorizações e totais comerciais devem ser validados pelo backend, mesmo quando exibidos ou calculados provisoriamente na interface.

## Sequência de implementação

1. Revisar a identidade minimalista, o catálogo demonstrativo e as telas do frontend.
2. Escolher a arte definitiva, substituir as camadas SVG provisórias e revisar o parallax.
3. Criar API com Fx, Fiber, Zap, conexão PostgreSQL, migrações, health checks, Swagger e Docker.
4. Entregar fluxo de catálogo completo: persistência → API → lista e produto no frontend.
5. Implementar fluxo comercial e administração conforme as decisões de produto.

O frontend está conectado ao catálogo Postgres via API Go. Contas e pedidos demonstrativos também são persistidos. Pagamento é simulado localmente; a loja não realiza cobranças ou entregas reais.

## Abstract Factory implementado

`application.RepositoryFactory` define a criação de `ProductRepository` e `TaxonomyRepository`. `postgres.Factory` cria implementações compatíveis sobre um único pool; `NewService` depende da fábrica abstrata, sem importar pgx. Fx seleciona e injeta a família Postgres na composição. Uma família de memória nos testes comprova a substituição.

Referência: [Abstract Factory](https://refactoring.guru/pt-br/design-patterns/abstract-factory). A família aqui é de serviços de persistência, não de entidades dos móveis. Entidades e regras continuam independentes de infraestrutura.

O schema inclui categorias, ambientes e produtos com acabamentos. Valores monetários são inteiros em centavos. Migrações embarcadas usam transação, lock e checksum. O catálogo inicial é demonstrativo.

O carregamento no layout SvelteKit pertence ao servidor e fornece dados por requisição, sem guardar o catálogo em estado global de SSR. As páginas e a validação das rotas usam os mesmos dados. A sacola local usa esses preços apenas para a simulação. Autorização, estoque e totais reais devem ser implementados antes de habilitar compras.

Detalhes de execução, rotas e testes: [API](../api/README.md).


## Fluxo comercial demonstrativo

`commerce.application.RepositoryFactory` cria repositórios de identidade e pedidos. Argon2id protege senhas de teste; sessões opacas são persistidas por hash e transportadas por um proxy SvelteKit com cookie HttpOnly. Preços, reserva de acabamento e total são validados dentro da operação do pedido.

`PaymentGateway` recebe uma implementação local injetada por Fx. Seus resultados explícitos de aprovação/recusa demonstram a transição `awaiting_payment → paid_demo`; cancelamento pendente usa `cancelled` e libera estoque. Não há chamadas a provedores externos. Confira contratos e limites em [API](../api/README.md).
