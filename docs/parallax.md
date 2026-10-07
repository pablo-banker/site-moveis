# Sala em camadas

A composição minimalista foi aprovada em 7 de outubro de 2026. A referência está em `docs/assets/parallax/sala-minimalista-base-v1.png`; o prompt original está em `docs/assets/parallax/conceito-v1.md`.

## Assets e profundidade

Todos os PNGs têm 1672 × 941 pixels. Gerados com a ferramenta integrada de imagens a partir da referência aprovada.

| Imagem em `web/static/background/` | Conteúdo | Profundidade |
| --- | --- | --- |
| room-background.png | Arquitetura e piso reconstruídos, sem móveis | 0,07 |
| room-rear.png | Luminária e mesa lateral, com alfa | 0,055 / 0,035 |
| room-sofa.png | Sofá inteiro e almofadas, com alfa | 0,04 |
| room-rug-table.png | Tapete e mesa de centro, com alfa | 0,025 |
| room-chair.png | Poltrona do primeiro plano, com alfa | 0,015 |

A luminária e a mesa lateral compartilham uma imagem, mas são renderizadas em dois planos com recortes CSS distintos para ajustar o registro. Portanto, são cinco arquivos e seis planos. Os recortes gerados variam ligeiramente da referência; escala e translação em `web/src/lib/parallax.ts` alinham a composição. O fundo foi reconstruído atrás dos objetos; sofá e tapete incluem áreas antes ocultas.

## Integração

`Parallax.svelte` é usado na página inicial, na história da marca e no sistema visual. Um único canvas proporcional mantém todas as camadas alinhadas; ResizeObserver ajusta o tamanho. Há 8% de folga para o movimento. No celular, o enquadramento prioriza o sofá abaixo do texto, com um gradiente claro para contraste.

A rolagem real da página dirige o movimento, inclusive por mouse, teclado ou toque. O hero começa montado com deslocamento zero. O deslocamento acompanha a saída completa do bloco, limitado à altura do próprio bloco. O movimento vertical foi reduzido: o fundo compensa 7% da rolagem, e a poltrona 1,5%. Os móveis também se deslocam lateralmente em direções e velocidades distintas, definidos por `lateral` em `parallax.ts`. O sofá, a poltrona e a mesa lateral vão para a esquerda, enquanto a mesa de centro vai para a direita. A mesa lateral usa coeficiente horizontal -0,12 para destacar seu movimento independente do sofá (-0,07). O deslocamento horizontal tem limite de 64 px e reduz proporcionalmente em telas menores, sem mudar o alinhamento inicial. O efeito deixa de ter o antigo limite de 240 px. Event listener passivo, requestAnimationFrame e IntersectionObserver limitam atualizações ao bloco visível. Observers, eventos e frames são limpos na desmontagem. Movimento reduzido desativa o efeito e suas mudanças são acompanhadas durante a visita.

## Prompts das camadas

Cada chamada recebeu a imagem aprovada como alvo de edição e exigiu câmera, proporção, coordenadas, iluminação e textura originais, sem zoom nem recentralização.

1. Fundo opaco: remover todos os móveis, objetos, tapete e sombras associadas; reconstruir arquitetura e piso atrás deles.
2. Plano distante transparente: extrair somente luminária completa e mesa lateral com livros; reconstruir a base oculta da luminária.
3. Sofá transparente: extrair sofá grafite e almofadas, reconstruindo bordas inferiores e lateral antes ocultas.
4. Conjunto próximo transparente: extrair tapete e mesa com livros e tigela; reconstruir as áreas do tapete cobertas pela poltrona e sofá, removendo suas sombras.
5. Primeiro plano transparente: extrair somente poltrona clara com contorno limpo e pequena sombra de contato.

Os SVGs geométricos anteriores permanecem como estudo histórico e não são usados pela composição atual.
