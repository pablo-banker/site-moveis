# Plano e implementação de movimento

Objetivo: dar vida ao visual minimalista mantendo leitura, navegação e formulários previsíveis.

## Sistema

| Movimento | Implementação | Ritmo |
| --- | --- | --- |
| Transição de página | Entrada CSS no conteúdo recriado ao mudar de rota | 280 ms |
| Entrada inicial e alternativa | Cascata iniciada imediatamente no conteúdo da primeira tela | 620 ms, sem espera de saída |
| Revelação na rolagem | IntersectionObserver e Web Animations API | 620 ms, deslocamento de 24 px |
| Cascata | Filhos diretos de grupos marcados | Intervalo 70 ms, atraso máximo 280 ms |
| Botões e cards | Resposta ao hover e clique | 180 ms |
| Diálogo e aviso | Opacidade e movimento curto | 180 a 300 ms |

Os tempos e curvas estão em `web/src/lib/styles/tokens.css`. Ações de revelação estão em `web/src/lib/actions/reveal.ts`; integração de navegação está no layout raiz.

## Aplicação

Cabeçalho e rodapé permanecem estáveis durante as trocas de rota. Apenas mudanças de caminho recebem transição de página: busca, filtros e hashes na mesma página não disparam saída/entrada geral.

Títulos, texto introdutório e ações usam cascata. Grades de produtos, ambientes e blocos de benefícios são revelados ao entrar na janela. Produto, formulários, sacola e resumo recebem entradas simples. Uma camada já revelada não repete a animação ao subir e descer a mesma página; novos cards adicionados por filtros são registrados automaticamente.

O conteúdo da primeira janela inicia sua cascata imediatamente durante a montagem, sem esperar o IntersectionObserver. Apenas elementos abaixo da dobra esperam a rolagem. O contêiner da página faz um deslocamento curto, sem um fade geral que possa esconder a cascata.

A navegação usa o layout com conteúdo identificado pelo caminho da rota. View Transitions API foi removida deste fluxo: os snapshots podiam mascarar o movimento dos componentes e introduzir uma sensação de pausa. Não há animação de saída bloqueando a próxima página.

## Acessibilidade e desempenho

- Sem interceptar a rolagem ou modificar sua velocidade.
- Conteúdo visível por padrão no HTML; a ocultação temporária só é ativada pelo mecanismo de revelação.
- `prefers-reduced-motion` desativa transições, cascatas, movimentos de hover e parallax.
- Mudanças da preferência em tempo real cancelam animações e mostram conteúdo pendente.
- Foco de teclado revela imediatamente o elemento que recebeu foco.
- Um IntersectionObserver e um MutationObserver por página; cleanup ao desmontar.
- Animações de entrada alteram apenas opacity e transform, sem alterar dimensões do layout.
- Navegação não aguarda animações; não depende da View Transitions API.

Referências: [onNavigate no SvelteKit](https://svelte.dev/docs/kit/$app-navigation#onNavigate), [View Transitions API](https://developer.mozilla.org/en-US/docs/Web/API/Document/startViewTransition).


## Rolagem suave

Lenis 1.3.26 no layout global suaviza a roda do mouse com interpolação 0,09 e RAF próprio. Mantém a posição de scroll nativa, utilizada pelo parallax e pelos observers. Âncoras usam offset de 24 px. Navegações cancelam o impulso anterior antes da restauração de scroll do SvelteKit. Menus/modais respeitam o overflow; áreas com rolagem própria continuam nativas. Touch e preferência por movimento reduzido mantêm scroll nativo; mudanças da preferência atualizam a instância. A instância e os listeners são destruídos ao desmontar.

Referência: https://github.com/darkroomengineering/lenis

Validação no navegador: roda do mouse ativa `lenis-smooth`; navegação para Móveis reinicia em scroll 0 e continua rolando; âncora da coleção inicia interpolação; busca aberta mantém scroll 0 e `lenis-stopped`. O bloqueio observa o overflow do body usado pelo Bits UI e pausa/retoma a instância. Check e build aprovados.
