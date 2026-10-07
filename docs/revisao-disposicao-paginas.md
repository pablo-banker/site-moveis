# Revisão da disposição das páginas

Data: 7 de outubro de 2026. Escopo: pesquisa de referências e avaliação da estrutura atual; nenhuma alteração de interface nesta revisão. Análise qualitativa, sem testes com compradores nem métricas de conversão. As referências externas foram consultadas em suas páginas públicas; a ordem extraída do conteúdo não confirma todos os detalhes visuais responsivos.

## Referências

- [Westwing](https://www.westwing.com.br/): busca, campanhas de ambientes, categorias, best-sellers, coleção e produtos. Alterna inspiração e compra; não exige uma longa história antes de qualquer produto.
- [Finnish Design Shop](https://www.finnishdesignshop.com/en-us/): busca e navegação comercial disponíveis, com um bloco Design Stories sobre casas, peças e designers. A narrativa funciona como caminho de descoberta.
- [IKEA — sala de estar](https://www.ikea.com/us/en/rooms/living-room/): categorias, composições de salas, ideias de decoração e referências aos produtos que aparecem nos ambientes. É a referência mais direta para a página de ambiente proposta.
- [MadeiraMadeira](https://www.madeiramadeira.com.br/): navegação ampla por categorias e ambientes, ofertas e acesso comercial. Serve de comparação para compra objetiva, mas sua densidade não combina tanto com a proposta minimalista da Forma.

## Diagnóstico da Forma

A home atual segue: hero com parallax → categorias → quatro produtos → destaque da sala de estar → informações de design, entrega e cuidados.

O hero apresenta a atmosfera pretendida, mas a passagem para o catálogo ocorre sem explicar a coleção. O destaque da sala vem depois dos produtos e tem pouca ligação com a seleção exibida. Sua rota `/ambientes/sala-de-estar` renderiza Catalog com filtro de ambiente, sem uma experiência editorial própria.

A busca no catálogo está bem posicionada antes de filtros e resultados. No cabeçalho, o ícone requer abrir um diálogo para digitar; tornar a busca visível no desktop facilitaria sua descoberta. Isso é uma hipótese de usabilidade para validar, não uma conclusão de teste. Não há necessidade de uma grande seção de busca interrompendo a narrativa da home.

A página `/ambientes` usa imagens e títulos adequados à descoberta, mas conta peças sem explicar o uso de cada espaço. O produto tem informações de compra, dimensões e cuidados bem próximas da decisão; sua descrição é igual para todas as peças e falta uma história específica após o bloco de compra.

As imagens dos ambientes e do parallax são ilustrativas e não correspondem necessariamente ao catálogo. Não prometer compra das peças fotografadas ou desenhar hotspots de produto sem estabelecer essa correspondência.

## Disposição proposta

1. Hero atual: atmosfera da coleção e CTA claro. Se o CTA continuar “Explore a coleção”, seu destino precisa mostrar uma coleção identificável; para catálogo genérico, usar “Ver móveis”.
2. Introdução curta à coleção: título, duas ou três frases sobre formas, materiais e uso, acompanhados de detalhe visual. Não inventar autoria, fabricação ou sustentabilidade.
3. Destaque da sala de estar: antecipar o bloco atual e conectar a narrativa ao cotidiano e à composição.
4. Produtos relacionados ao ambiente/coleção apresentado: seleção coerente, com preço e acesso direto ao produto.
5. Categorias: atalhos para quem quer escolher por tipo, complementados pelo menu e busca sempre acessíveis.
6. Outros ambientes: poucas entradas visuais para ampliar a descoberta.
7. Confiança: entrega, montagem, trocas e cuidados com conteúdo concreto conforme a operação for definida.

Na página de ambiente: imagem ampla → pequena explicação da proposta → composição e detalhes → peças relacionadas → catálogo filtrado. Pode haver um link “Ver peças” no topo para pular à seleção.

Na página de catálogo: manter busca, filtros e ordenação no começo; não colocar uma longa narrativa antes dos resultados. Filtros de medida, preço e material ficam para a evolução do catálogo real.

Na página do produto: preservar preço, acabamento, entrega e ação de compra no começo. Depois apresentar uso, materiais, detalhes e combinação com outras peças, com conteúdo próprio de cada produto.

## Prioridades

1. Criar uma página de ambiente que cumpra a promessa de “Explore a sala de estar”.
2. Reordenar a home para conectar coleção, ambiente e seleção de produtos.
3. Tornar a busca mais fácil de identificar no desktop, mantendo o catálogo direto.
4. Substituir textos genéricos por conteúdo específico e aproximar imagens e catálogo.

Recomendações de projeto inferidas da comparação; não há evidência nesta pesquisa de aumento de conversão.

## Implementação após aprovação

Aplicada a sequência hero → introdução da coleção → sala de estar → seleção da sala → categorias → outros ambientes → informações de confiança. O CTA da coleção aponta para a introdução na home. As quatro páginas de ambientes agora têm narrativa, imagem, dicas e catálogo incorporado, com atalho para peças. A busca está visível no cabeçalho a partir de 1051 px; em telas menores permanece o diálogo. As descrições dos oito produtos são próprias; recomendações relacionadas usam o mesmo ambiente e ficam ocultas quando não há peças complementares. Imagens continuam como referências, sem hotspots nem promessa de correspondência exata.

Verificados no navegador: navegação da home para a sala, atalho às quatro peças da sala, filtro de poltronas, busca “sofá” no cabeçalho e navegação produto → ambiente. Layouts conferidos em 1280 px e 390 px.
