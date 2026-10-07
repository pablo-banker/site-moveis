# Revisão de segurança — 7 de outubro de 2026

A revisão abrangeu o frontend SvelteKit, BFF, API Go, autenticação, autorização, pedidos, estoque, cotações, pagamento, contato, dependências e configuração Docker. Foram corrigidos os problemas abaixo e executados testes locais. Ainda existem requisitos de lançamento: o projeto não deve ser declarado seguro para operação pública apenas pelo resultado dos scanners.

## Achados corrigidos

| Prioridade | Achado | Correção |
| --- | --- | --- |
| Alta | Toolchain Go antigo com vulnerabilidades conhecidas alcançáveis | Go e builder Docker atualizados para 1.26.8; dependências atualizadas. A versão do binário efetivamente instalado foi conferida. |
| Alta | Processamento de pagamento ocorria antes de bloquear o pedido | Autorização, bloqueio da linha e verificação do estado precedem o provedor. Um pedido já pago retorna seu estado sem processar novamente; cancelado ou vencido é rejeitado. |
| Alta | Proxy recebia o corpo inteiro antes de limitar tamanho | JSON limitado a 32 KiB por bytes, inclusive em transferência chunked, com prazo de leitura de 8 segundos; timeout de upstream e redirecionamentos proibidos. API também limita tamanho/tempo. |
| Alta | Custo de Argon2 e tentativas por conta sem proteção compartilhada | Máximo de dois cálculos Argon2 simultâneos por processo; 8 tentativas por minuto por e-mail normalizado com atualização atômica no Postgres. Novos cadastros exigem 15 caracteres, até 128 bytes. |
| Média | Reservas de estoque podiam ficar pendentes indefinidamente | Vencimento de 30 minutos; tarefa a cada minuto cancela e devolve estoque em transação, com bloqueio e `SKIP LOCKED`. Pagamento confere o relógio atual após obter o bloqueio. |
| Média | Ausência de CSP e cabeçalhos de proteção no site | CSP com nonce por resposta SSR, bloqueio de frames/objetos e origens explícitas. `nosniff`, política de referer e permissões restritas. |
| Média | Produção podia herdar configurações inseguras de desenvolvimento | API valida ambiente, TLS Postgres `verify-full`, senha não padrão, CORS HTTPS explícito e migrations fora do processo. Frontend exige origem HTTPS explícita e API configurada. |
| Média | Cookie de sessão e configuração de origem dependiam apenas da URL recebida | Produção usa `__Host-forma-session`, HttpOnly, Secure obrigatório, SameSite=Lax e path `/`; mutações conferem Origin e rejeitam Fetch Metadata cross-site. |
| Média | Falha em uma tentativa de login apagava a sessão existente | Login/cadastro malsucedidos preservam o cookie; logout e sessão inválida em rotas autenticadas continuam revogando/limpando normalmente. |
| Baixa | Dependência transitiva `cookie` afetada por CVE-2024-47764 | Override para 0.7.2 e lockfile atualizado; auditoria npm sem alertas. |
| Defesa adicional | Swagger público por padrão, permissões do container e registros expirados | Docs desabilitados por padrão em produção; API não root, filesystem somente leitura, sem capabilities, sem novos privilégios e com limites de memória/processos. Limpeza periódica limitada de sessões, cotações e tentativas antigas. |

A CSP permite estilos inline para as transições/animações existentes, mas não permite scripts inline sem nonce. Swagger de desenvolvimento tem política própria para sua CDN. HSTS só é emitido em produção com origem HTTPS validada.

A migration de expiração também define o prazo de pedidos antigos a partir da criação. Pedidos pendentes antigos podem ser cancelados pela manutenção; pedidos pagos não são afetados.

## Controles existentes conferidos

- SQL parametrizado e lista permitida de ordenações; entradas validadas e JSON com campos desconhecidos rejeitados.
- Autorização por proprietário para consultar, cancelar e pagar pedidos; IDs de outra conta não revelam o pedido.
- Preços, totais e estoque calculados/conferidos no servidor. Frete persistido, vinculado a cliente, CEP e itens, com prazo e snapshot no pedido.
- Criação idempotente e reserva transacional; saldo insuficiente reverte a operação inteira.
- Argon2id com salt aleatório e comparação constante; tokens aleatórios de 256 bits, apenas hash no banco, expiração e revogação. O BFF remove o token do JSON e o mantém em cookie HttpOnly.
- Nenhum token de sessão em localStorage; o armazenamento local contém apenas sacola e favoritos. Não foram encontrados `@html`, `innerHTML` ou `eval` no código da aplicação.
- Logs HTTP não registram corpos, cookies ou Authorization; falhas internas não retornam detalhes de banco ao cliente.
- Consulta de CEP usa host fixo HTTPS, exatamente oito dígitos, timeout e não segue redirecionamentos. Não encaminha cookies do cliente ao ViaCEP.
- `.gitignore` e `.dockerignore` excluem `.env`; apenas exemplos foram encontrados no inventário. Não há repositório Git neste diretório, portanto não foi possível revisar histórico de segredos.

## Evidências da validação

- `go test -race ./...` com Postgres real em schemas isolados: aprovado; `go vet ./...`: aprovado.
- Concorrência: cinco pagamentos simultâneos do mesmo pedido chamaram o provedor uma vez; pagamento de outra conta foi rejeitado; pagamento vencido foi rejeitado; duas execuções de manutenção restituíram estoque uma vez; 24 tentativas concorrentes admitiram exatamente oito no limitador por conta.
- Configuração: testes rejeitam senha padrão, banco sem TLS validado, CORS aberto/HTTP e migrations automáticas em produção.
- `npm run check`: zero erros e avisos; `npm run build`: aprovado. Adapter Node configurado e imagem de produção compilada.
- `npm audit`: zero vulnerabilidades reportadas.
- `govulncheck` no código e no binário Docker Go 1.26.8: zero vulnerabilidades em símbolos usados/pacotes importados. Existe um aviso de módulo para `golang.org/x/crypto/openpgp`, que a loja não importa. O projeto usa `argon2` desse módulo. O aviso não tem versão corrigida; deve continuar sendo monitorado. [Aviso oficial GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932).
- O build preserva símbolos para permitir análise do artefato; sem eles, o scanner apresentava o aviso de OpenPGP conservadoramente como possível uso.
- `npm run security:smoke` contra a loja local: cabeçalhos/CSP, nonce diferente por resposta, CSRF, JSON obrigatório, tamanho por bytes/chunks, allowlist de rotas, cache privado e preservação de sessão em login incorreto aprovados. Upload conhecido excedente retorna 413; cancelar stream chunked pode encerrar a conexão no adapter Node, também sem encaminhar o corpo para a API.
- API reconstruída e migrations aplicadas no ambiente local. O smoke de produção usou certificados verificados por uma CA de testes, SMTP STARTTLS obrigatório e Postgres TLS, em containers isolados. Não houve publicação em infraestrutura pública ou processamento financeiro real. A tela de recuperação também foi conferida no navegador.

## Requisitos de publicação aplicados nesta etapa

| Controle | Implementação e evidência |
| --- | --- |
| Infraestrutura | Compose de produção separado; adapter Node; Caddy com HTTPS; API privada; imagens com digests fixados; credencial de migration separada. Runtime rejeita usuário administrativo, dono do banco ou com CREATE no schema. |
| Abuso compartilhado | Quotas atômicas no Postgres por IP/conta; Caddy sobrescreve IP e BFF assina método/caminho/IP/timestamp; assinatura inválida/expirada é rejeitada. CEP tem cache persistido, timeout e concorrência limitada. |
| Contas | Cadastro genérico 202 sem sessão automática; confirmação de e-mail antes do checkout; recuperação com token único/expiração; revogação de sessões e painel para encerrar outras sessões. |
| E-mails | Outbox transacional com AES-GCM, worker com retry e limite de tentativas; SMTP com validação de certificado em produção. Tokens nos fragmentos dos links e páginas de conta com no-referrer. |
| Operação | Métricas privadas com Bearer, regras Prometheus, templates de alertas e timer de backup. Backup age criptografado e restauração em banco vazio testados. Eventos de confirmação, redefinição e revogação ficam registrados. |
| Dependências | Pipeline de testes/auditorias preparado. Trivy nas imagens finais da API, web e backup: zero HIGH/CRITICAL na verificação de 07/10/2026. Removidas ferramentas npm e gosu desnecessárias que traziam vulnerabilidades ao runtime. |

O teste `scripts/deployment_smoke.py` validou HTTPS/HSTS/CSP, Postgres TLS com usuário restrito, HMAC BFF, cookie Secure/HttpOnly, cadastro → SMTP → confirmação → login → frete → pedido → pagamento pelo adaptador do servidor, recuperação com invalidação da sessão antiga, métricas e backup/restauração. Os containers e dados desse teste foram descartados.

Testes adicionais verificaram cadastro duplicado com mesma resposta pública, ciphertext da outbox, checkout de conta não confirmada rejeitado, consumo concorrente de token com exatamente um sucesso, finalidade/expiração dos tokens, revogação de outras sessões e quotas compartilhadas concorrentes. `go test -race`, `go vet`, check/build/audit npm e govulncheck foram repetidos após as alterações e passaram.

A configuração ainda precisa de domínio/hospedagem/SMTP reais, segredos e credenciais exclusivos, ativação do timer, cópia externa dos backups e receptor de alertas. Os templates não executam esses serviços no host automaticamente. Políticas de retenção/exclusão dos dados comerciais, catálogo licenciado, atendimento e eventual acesso administrativo continuam dependendo da operação. Veja [guia de publicação](publication.md) e [limites funcionais](production-readiness.md).

Pagamento/envio reais estão fora desta etapa. Para habilitá-los futuramente, será necessário implementar também eventos assíncronos, reconciliação e expedição; não basta trocar a configuração do adaptador síncrono atual.

## Repetir verificações

Na raiz:

```sh
npm --prefix web run check
npm --prefix web run build
npm --prefix web audit
npm --prefix web run security:smoke
```

O smoke pressupõe API/site locais ativos; `SECURITY_TEST_URL` permite outro endereço de testes. Use exclusivamente um ambiente autorizado.

Em `api`, com `TEST_DATABASE_URL` apontando para um banco de testes:

```sh
go test -race ./...
go vet ./...
govulncheck -show verbose ./...
```

Confirme o toolchain efetivamente utilizado com `go version`; repetir com um toolchain diferente pode mudar os resultados. A análise do artefato também pode usar `govulncheck -mode=binary -show verbose CAMINHO_DO_BINARIO`.

Referências: [OWASP Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html), [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html), [Go Vulnerability Management](https://go.dev/doc/security/vuln/), [advisory cookie](https://github.com/advisories/GHSA-pxg6-pf52-xh8x).
