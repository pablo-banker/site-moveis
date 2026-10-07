# Verificação do frontend

Data: 6 de outubro de 2026.

Fluxo revisado: início → catálogo → busca por sofá → produto → acabamento e quantidade → sacola → checkout demonstrativo → confirmação.

## Evidências de navegador

- Busca por “sofá”: 1 móvel encontrado, Sofá Arco.
- Produto: acabamento Grafite e quantidade 2 mantidos na sacola.
- Sacola: subtotal de R$ 8.580,00; seleção preservada depois de recarregar.
- Checkout: dados fictícios, entrega e seleção de Pix disponíveis em três etapas.
- Confirmação: total simulado de R$ 8.729,00, incluindo R$ 149,00 de frete; sacola esvaziada.
- CEP: corrigida a interpolação indevida de quantificadores do atributo pattern; CEP com 8 dígitos e hífen passou na validação do navegador.
- Celular, viewport de 390 × 844: menu abriu e navegou para o catálogo; filtros expandiram; largura do conteúdo permaneceu dentro da janela.
- Parallax: as três camadas mudaram suas translações após rolar, chegando a 14,4 px, 28,8 px e 48 px no ponto observado.
- Proteção de movimento reduzido: implementada em CSS e no listener de media query; a alternância da preferência do sistema não foi testada manualmente.

## Limites

Não existe API nesta fase. O fluxo usa catálogo estático, armazenamento local para sacola e favoritos e armazenamento de sessão para confirmação. Fotos externas, arte final do fundo, conteúdo comercial e marca definitiva continuam sujeitos a revisão. Administração e operações reais da conta são fases futuras.

## Movimento — revisão adicional

- Svelte check: 0 erros e 0 avisos; build de produção concluído.
- Início: 7 grupos de cascata registrados. Após rolar para a seleção, todos os quatro cards ficaram visíveis, com opacity 1 e sem a classe de estado pendente.
- Navegação início → catálogo → início concluiu e restaurou a opacidade do conteúdo.
- Busca por “mesa” após a navegação continuou retornando 3 produtos.
- Diálogo de busca abriu e fechou em viewport de 390 × 844, com foco no botão de fechar.
- O mecanismo respeita movimento reduzido por CSS e media query; a mudança da preferência do sistema não foi testada manualmente.


## Backend — primeira etapa (07/10/2026)

- Go: testes com detector de corrida e `go vet` aprovados. Testes Postgres usam schema temporário isolado, removido ao terminar.
- Migrações: schema e oito produtos demonstrativos; repetição não duplica registros nem sobrescreve preços.
- HTTP: busca, categoria, ambiente, ordenação, paginação, 400 em filtro inválido, 404 em produto inexistente, request ID e health checks validados.
- Persistência: alteração no banco aparece na API. No navegador, preço do Sofá Arco alterado temporariamente de R$ 4.290,00 para R$ 4.290,99 apareceu no catálogo e foi restaurado.
- Navegador: catálogo, detalhe e inclusão/remoção na sacola funcionando com dados da API. Sacola restaurada vazia.
- Disponibilidade: API parada retorna 503 no frontend; reinício recupera o catálogo. Há uma página estática com a identidade da loja para falhas do carregamento raiz.
- Docker: imagem multi-stage construída; API e Postgres iniciados. O helper de credenciais do Docker Desktop bloqueou a primeira construção. A segunda usou configuração temporária anônima para imagens públicas, sem modificar credenciais do usuário.
- Swagger UI: abre `/docs` e carrega o contrato local, com os quatro endpoints de catálogo e dois de saúde.
- Svelte: check sem erros/avisos e build aprovado após integrar o catálogo.

Busca e filtros da interface ainda operam localmente sobre o catálogo recebido; a API também oferece filtros paginados. Autenticação, estoque comercial, pedidos, frete e pagamentos permanecem para as próximas etapas.


## Contas e compra demonstrativa — 07/10/2026

- Testes de integração com schema temporário: cadastro/login, proteção da senha, token revogado/expirado, acesso bloqueado a pedidos de outra conta, reserva idempotente, rollback integral e cancelamento devolvendo estoque uma vez.
- Concorrência: seis pedidos para cinco unidades; cinco reservas aceitas e uma recusada, sem estoque negativo.
- Pagamento local: recusa mantém pendente; aprovação muda para paid_demo e repetição não duplica efeitos; pedido cancelado não pode ser pago, aprovado não pode ser cancelado.
- Navegador: conta fictícia criada, catálogo mostra cinco unidades por acabamento, checkout cria pedido e limpa a sacola; confirmação demonstra recusa seguida de aprovação sem cobrança. Pedido de apresentação pertence à conta demo-navegador@example.com, com dados fictícios.
- Próximos recursos, caso o projeto evolua: recuperação de senha, e-mail de confirmação, expiração de reserva e integração externa sandbox. Não necessários para a apresentação local atual.


## Consulta de CEP no checkout

O checkout consulta o ViaCEP por uma rota SvelteKit `/endereco/[cep]`, enviando somente o CEP ao provedor. A rota valida os oito dígitos, tem timeout de cinco segundos e trata CEP inexistente ou indisponibilidade. A interface preenche rua, cidade e UF, preserva edições durante a consulta e cancela respostas obsoletas ao mudar o CEP. Campos continuam editáveis; CEP genérico exige a rua manualmente. O número e o frete demonstrativo não são calculados pelo serviço.

Validação: `npm run check --prefix web` e `npm run build --prefix web` aprovados. Consulta no navegador com 88301401 retornou Rua Lauro Muller, Itajaí, SC.


## Frete e fluxo de loja (07/10/2026)

Testes Go com race detector e vet passaram após migrations 005–007. Incluem cotação por CEP/quantidade, rejeição de cotação de outro cliente, CEP diferente, expiração e campos de frete fornecidos pelo cliente; retry idempotente após expiração, rollback de estoque e rejeição de outcome de pagamento enviado pelo navegador. Contato valida e persiste mensagens. Check e build Svelte passaram sem erros ou warnings.

No navegador, produto e checkout cotaram 88301401 em R$ 189,00 / 10–15 dias; subtotal R$ 4.290,00 e total R$ 4.479,00. A confirmação existente exibe Pagar pedido sem controles de simulação. Nenhum novo pedido ou pagamento foi feito nessa validação visual.
