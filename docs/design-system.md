# Produto e identidade visual

Status: loja com carrinho e pagamento online e direção minimalista escolhidos pelo usuário. Frontend demonstrativo implementado; backend será iniciado após revisão visual. Nome da marca e imagem definitiva do fundo ainda não foram definidos.

## Público e objetivo

Hipótese de trabalho: pessoas que buscam móveis residenciais contemporâneos e precisam comparar dimensões, materiais, acabamentos e condições de entrega antes de comprar ou solicitar orçamento.

A navegação deve permitir descobrir um móvel por categoria ou ambiente, verificar se cabe no espaço e avançar para o fluxo comercial escolhido. Nome, logotipo e textos institucionais serão definidos com a marca real.

## Direção visual escolhida

Branco, grafite e cinza. Fotografias amplas, títulos sem serifa, hierarquia clara e espaço entre seções. Forma é apenas uma marca provisória de estudo.

| Cor | Hex | Papel |
| --- | --- | --- |
| Branco | `#FFFFFF` | Superfícies e conteúdo |
| Grafite | `#242424` | Ações e texto principal |
| Névoa | `#F7F7F5` | Fundos secundários |
| Cinza | `#666666` | Texto secundário |
| Cinza suave | `#A5A5A0` | Detalhes decorativos |
| Borda | `#E2E2DE` | Divisórias decorativas |

Estados usam sucesso `#246347`, aviso `#795600` e erro `#B42318`. Campos usam borda com contraste próprio `#858580`. Estados nunca devem depender apenas de cor.

## Tipografia e medidas

Helvetica Neue, Helvetica e Arial como fontes locais de sistema, sem requisições externas. Títulos usam peso 500 e espaçamento discreto entre letras; conteúdo de leitura permanece em 15 px, com metadados menores.

- Texto base: 15 px; entrelinha 1,5.
- Escala: 12, 14, 16, 20, 28, 40 e 56 px; títulos grandes fluidos em telas pequenas.
- Espaçamento: 4, 8, 12, 16, 24, 32, 48, 64 e 96 px.
- Conteúdo: largura máxima de 1280 px; margens de 16 px no celular e 32 px em telas maiores.
- Leitura: blocos de texto com até 70 caracteres por linha.
- Cantos: 12 px em campos e controles, 16 px em imagens, 24 px em cards e painéis; botões e chips em formato de pílula. Raios centralizados nos tokens globais.
- Layout alinhado à esquerda; fotos amplas e proporções consistentes no catálogo.

## Páginas comuns aos modelos comerciais

| Página | Rota proposta | Conteúdo principal |
| --- | --- | --- |
| Início | `/` | Ambiente em destaque, categorias, seleção de móveis e informações de entrega |
| Catálogo | `/moveis` | Busca, filtros por categoria, ambiente, material, dimensões e preço quando público |
| Categoria | `/moveis/categoria/[slug]` | Catálogo contextual com os mesmos filtros |
| Produto | `/moveis/[slug]` | Galeria, medidas, materiais, variantes, disponibilidade e ação comercial |
| Ambientes | `/ambientes` | Inspiração para sala, quarto, jantar e escritório |
| Ambiente | `/ambientes/[slug]` | Composição e produtos relacionados |
| Sobre | `/sobre` | Marca, fabricação e materiais |
| Contato | `/contato` | Canais e horários reais de atendimento |
| Ajuda | `/ajuda` | Entrega, montagem, cuidados e dúvidas |
| Privacidade | `/privacidade` | Conteúdo a definir conforme as operações reais |
| Termos | `/termos` | Conteúdo a definir conforme o modelo comercial |
| Sistema visual | `/design-system` | Página de revisão de tokens, componentes e parallax |

O produto deve apresentar largura, altura e profundidade com unidades explícitas, prazo de produção separado de entrega e variantes com nomes de acabamentos. Dimensões e prazo têm prioridade perto da ação principal.

## Fluxo comercial escolhido

Loja online: catálogo → produto e acabamento → carrinho → identificação → entrega → pagamento → confirmação.

Telas implementadas: `/carrinho`, `/checkout`, `/pedido-confirmado`, `/conta` e `/conta/pedidos`. O checkout é uma simulação local com frete demonstrativo; não solicita dados de cartão. Conta e cadastro são apenas prévias visuais.

Pagamento, transportadora e autenticação reais serão definidos na etapa do backend.

## Administração proposta

Rotas em `/admin`, com autenticação e autorização próprias: produtos, categorias, ambientes, variantes, imagens e disponibilidade. Pedidos só entram no modelo de loja; orçamentos e projetos entram conforme o modelo escolhido. Nunca confiar no bloqueio visual do frontend como autorização.

## Componentes iniciais

Cabeçalho, navegação móvel, busca, breadcrumb, cartão de produto, filtros, galeria, seletor de acabamento, bloco de medidas, botão, campo de formulário, diálogo e aviso de estado. Cada componente deve prever carregamento, vazio, erro, indisponibilidade e foco de teclado.

Proposta: Bits UI para comportamentos como diálogo, seleção e accordion; Tailwind para composição visual; componentes próprios em `src/lib/components/ui` para aplicar os tokens de forma consistente. Bits UI permite estilização própria, conforme sua [documentação](https://www.bits-ui.com/docs).

## Composição da página inicial

```text
Logo        Móveis   Ambientes   Sobre       Busca   Ação comercial
─────────────────────────────────────────────────────────────────
Título direto + ação       Foto ampla de um ambiente mobiliado
─────────────────────────────────────────────────────────────────
Categorias: Sofás | Mesas | Cadeiras | Estantes
Seleção de produtos: foto, nome, dimensões, preço quando público
Ambiente em destaque com produtos identificados
Entrega, montagem e atendimento com informações reais
Rodapé: contato, ajuda e páginas institucionais
```

Celular: navegação compacta, conteúdo em coluna, catálogo em uma ou duas colunas conforme espaço disponível e filtros em diálogo. Evitar carrossel automático e animação decorativa.

## Paleta global

Fonte única: `web/src/lib/styles/tokens.css`, importada por `web/src/app.css` no layout raiz do SvelteKit. Componentes não devem redefinir a paleta ou duplicar valores hexadecimais. Cores de fotos e amostras de acabamento de produtos são dados do catálogo, não tokens da interface.

Uso em CSS: `background: var(--color-primary); color: var(--color-on-primary);`.

Integração com Tailwind CSS 4 presente no `app.css`:

```css
@import 'tailwindcss';
@import './lib/styles/tokens.css';

@theme inline {
  --color-primary: var(--ui-primary);
  --color-on-primary: var(--ui-on-primary);
  --color-background: var(--ui-background);
  --color-surface: var(--ui-surface);
  --color-foreground: var(--ui-foreground);
  --color-muted: var(--ui-muted);
}
```

Exemplo de classes: `bg-primary text-on-primary`, `bg-surface text-foreground`. Os demais tokens podem ser consumidos diretamente via CSS custom properties. O uso de `@theme inline` segue a [documentação oficial](https://tailwindcss.com/docs/theme).

## Critérios de revisão

Contraste mínimo de 4,5:1 para texto comum e 3:1 para texto grande; foco de teclado visível; rótulos permanentes nos formulários; nenhuma informação importante exclusivamente em hover; respeito a movimento reduzido; fotos com descrição útil; ausência de rolagem horizontal no celular. O token de borda decorativa não deve ser a única indicação de um campo: usar o token de borda de controle.

## Decisões pendentes

Nome da marca; público e faixa de preço; catálogo real; regiões atendidas; regras de entrega e montagem; fotos e conteúdo; provedor de pagamento. A administração é uma fase futura, ainda não implementada.

## Fundo parallax

Sala minimalista aprovada, separada em cinco imagens PNG e seis planos de movimento. Aplicada na página inicial, em `/sobre` e `/design-system`. Ver [camadas e movimento](parallax.md).

## Identidade editorial ampliada

A coleção usa um capítulo em grafite com tipografia de grande escala, imagem em recorte de arco e um estágio que acompanha a leitura em telas maiores. O destaque da sala usa fotografia quase em tela inteira, com texto sobreposto. As páginas de ambientes e a história da marca recebem títulos e áreas de imagem mais amplos. O catálogo mantém escala funcional.

`web/src/lib/actions/editorial.ts` fornece progresso de rolagem às imagens e à assinatura tipográfica. Os efeitos usam transformações, evento passivo e requestAnimationFrame apenas enquanto o bloco está visível. No celular a coleção não fixa a rolagem; movimento reduzido remove a fixação e os deslocamentos. Mantidas a paleta branca/grafite/cinza, formas arredondadas e composição parallax aprovada.

Verificações: build e svelte-check aprovados, revisão visual em 1280 × 900 e 390 × 844, sem overflow horizontal no capítulo editorial e sem erros no console durante a revisão da home.

## Aplicação às demais páginas

Ambientes usa abertura com imagem em arco e composições grandes em duas colunas alternadas. O catálogo recebe título de grande escala e uma barra de busca e ordenação arredondada, mantendo resultados e filtros diretos. Produtos têm galeria fixa no desktop, zoom suave ligado ao scroll e história em um painel grafite. Conta combina um bloco de identidade com o formulário; sacola e checkout mantêm a hierarquia de compra com resumos mais definidos. Contato, ajuda, termos, privacidade e confirmação compartilham a tipografia e os painéis da identidade.

A navegação de contato foi corrigida: a correspondência de conta é exata para `/conta` e `/conta/pedidos`, evitando capturar `/contato`.

Validação funcional: busca por sofá, abertura de produto, adição à sacola, entrada no checkout, retirada do item de teste, contato e expansão da resposta de entrega na ajuda. Checkout móvel sem overflow horizontal. A sacola foi restaurada ao estado vazio após o teste.


## Área do cliente

`/conta` funciona como visão geral: dados do cadastro, endereço informado no último pedido, atalhos para histórico e favoritos locais e resumo de até dois pedidos recentes. Não inclui ações de cancelamento ou pagamento.

`/conta/pedidos` concentra o histórico retornado pela API (até 100 pedidos recentes), itens, status, totais e acesso aos detalhes/simulação de pagamento. O filtro distingue todos, pendentes, aprovados e cancelados; cancelamento aparece apenas para pedidos pendentes. A navegação indica a página ativa e os breadcrumbs distinguem as rotas.

Endereço é identificado como dado do último pedido, pois ainda não existe um cadastro independente de endereços. Dados pessoais são apresentados para consulta.
