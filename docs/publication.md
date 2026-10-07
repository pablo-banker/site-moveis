# Publicação da Forma

Configuração preparada para Docker Compose em um host Linux, Caddy como entrada HTTPS e Postgres externo com TLS. Hospedagem, domínio e SMTP ainda não foram escolhidos. Este guia não publica nada automaticamente.

## 1. Preparar o destino e os segredos

Disponibilize Docker/Compose, acesso restrito ao host, domínio apontando para seu IP e portas 80/443. API, web, banco e métricas não devem receber portas públicas adicionais. O Compose de desenvolvimento não serve para publicação.

Na raiz do checkout destinado à publicação:

```sh
cp deploy/.env.production.example deploy/.env.production
chmod 600 deploy/.env.production
```

Preencha esse arquivo sem colocá-lo no Git. Para `BFF_SHARED_SECRET` e `METRICS_TOKEN`, gere dois valores independentes com `openssl rand -hex 32`. Para `EMAIL_OUTBOX_KEY`, use `openssl rand -base64 32`. Grave a chave da outbox no gerenciador de segredos, junto às credenciais; ela será necessária para recuperar mensagens pendentes de um backup. Não substitua a chave enquanto existirem mensagens pendentes sem um plano de rotação.

| Variável | Configuração |
| --- | --- |
| SITE_HOST | Hostname sem protocolo/caminho; Compose deriva SITE_ORIGIN/ORIGIN HTTPS |
| TLS_EMAIL | E-mail operacional para certificados Caddy |
| DATABASE_URL | Credencial runtime restrita, `sslmode=verify-full` |
| MIGRATION_DATABASE_URL | Credencial separada com propriedade/DDL dos objetos |
| RUNTIME_DB_ROLE | Nome do usuário runtime existente |
| BFF_SHARED_SECRET | Mesmo segredo na API e no BFF, mínimo 32 caracteres |
| METRICS_TOKEN | Segredo independente para métricas privadas |
| EMAIL_OUTBOX_KEY | Exatamente 32 bytes em base64 |
| SMTP_HOST/PORT | Endpoint do serviço escolhido; normalmente 587 ou 465 |
| SMTP_TLS_MODE | `starttls` para STARTTLS obrigatório ou `tls` para TLS implícito |
| SMTP_USER/PASSWORD | Credenciais exigidas pelo serviço |
| SMTP_FROM | Remetente autorizado, por exemplo `Forma <conta@dominio>` |
| BACKUP_DATABASE_URL | Usuário dedicado com leitura dos objetos necessários |
| BACKUP_AGE_RECIPIENT | Chave pública age; chave privada fora do servidor |
| BACKUP_DIRECTORY_HOST | Diretório absoluto persistente no host de produção |

Senhas nas URLs devem ser codificadas para URI. A cadeia da CA do Postgres/SMTP precisa ser reconhecida pelos containers. Se o banco usa CA privada, monte o certificado somente leitura nos serviços api/migrate/backup e inclua `sslrootcert=/caminho/ca.crt` nas URLs; nunca desative verificação de certificado. Configure o remetente/domínio no SMTP escolhido antes de testar envio externo.

## 2. Provisionar Postgres e executar migrations

Use um banco novo para produção. Não copie o volume/credenciais de desenvolvimento. Um operador provisiona usuários com segredos próprios:

- `forma_migrator`: proprietário dos objetos da aplicação, com DDL para migrations.
- `forma_runtime`: sem superuser/CREATEDB/CREATEROLE, sem propriedade do banco, sem CREATE no schema e sem participação em papéis administrativos.
- `forma_backup`: leitura dos objetos necessários a pg_dump; sem escrita/DDL, acesso de rede restrito.

O migrator precisa ter controle sobre o schema `public` para conceder permissões. O CLI revoga CREATE do PUBLIC, concede CRUD/sequências ao runtime e retira acesso à tabela `schema_migrations`. Não concede permissões ao backup automaticamente: o operador deve conceder CONNECT, USAGE no schema e SELECT nas tabelas/sequências, incluindo novas tabelas após cada migration, ou usar defaults do proprietário adequados ao papel de backup.

```sh
docker compose --env-file deploy/.env.production -f compose.production.yaml build api web backup
docker compose --env-file deploy/.env.production -f compose.production.yaml run --rm migrate
docker compose --env-file deploy/.env.production -f compose.production.yaml up -d --wait caddy web api
```

Não rode `compose config` em logs públicos: a saída expandida inclui segredos. `AUTO_MIGRATE=false` e docs desabilitados são impostos pela configuração de produção. Migrações não têm rollback destrutivo automático. Faça backup antes de atualizações e confira compatibilidade do schema com uma imagem anterior antes de reverter binários.

Seeds atuais são referências de catálogo/estoque. Revise-os e use dados/fotos autorizados antes de abrir a loja. Caddy mantém certificados nos volumes; preserve esses volumes nas atualizações. Com DNS e portas corretos, solicita certificados públicos. O IP encaminhado é sobrescrito por Caddy, o adapter confia em um salto e o BFF assina as chamadas da API. Se adicionar CDN/proxy, revise explicitamente a cadeia de confiança antes de usar IP de headers adicionais.

## 3. Conferir contas e envio de e-mail

Cadastre uma conta de teste controlada, confirme o e-mail e entre na conta. Confirme que pedidos são recusados antes da verificação. Verifique reenvio, link expirado, recuperação de senha, invalidação das sessões antigas e encerramento de outras sessões. Links de confirmação/reset usam fragmentos e a confirmação depende de ação explícita do usuário.

Cadastro e recuperação retornam mensagens genéricas para não revelar existência de uma conta. A API grava desafio e e-mail na transação; o worker envia a cada dez segundos e tenta novamente até dez vezes. Falhas aparecem nas métricas sem registrar senha/token/corpo. Uma fila com tentativas esgotadas exige intervenção e nova solicitação pelo usuário após corrigir o SMTP.

Em desenvolvimento, `docker compose up -d --build` disponibiliza Mailpit em `http://127.0.0.1:8025`; mensagens não saem para serviços externos. A chave padrão da outbox e SMTP sem TLS são permitidos apenas no desenvolvimento. Produção rejeita a falta de TLS, origem HTTPS ou chave explícita.

## 4. Ativar backups e testar restauração

Gere uma identidade com `age-keygen` em uma máquina segura. Guarde a chave privada offline/em cofre e use somente o recipient público no servidor. Configure usuário de leitura e diretório persistente de backup. O dump sai criptografado pelo pipeline, com arquivo temporário removido em falhas e renomeado ao concluir; não há dump plaintext persistido.

```sh
docker compose --env-file deploy/.env.production -f compose.production.yaml run --rm backup
```

Instale os templates `deploy/backup/forma-backup.service` e `.timer` no host, ajustando `WorkingDirectory=/opt/forma` ao checkout. Um operador ativa o timer com `systemctl enable --now forma-backup.timer`. Ele roda diariamente às 03:00 UTC. Esses arquivos são templates: nada foi instalado/agendado no host nesta etapa.

Configure cópia dos arquivos criptografados para armazenamento externo, retenção, alarme para ausência de backup recente e teste periódico de restauração. O script não faz upload nem apaga backups antigos. Cópia no mesmo disco não cobre perda do host.

Restaure somente em um banco novo e isolado. Monte o arquivo criptografado e a identidade privada somente leitura no container de backup; defina `RESTORE_DATABASE_URL`, `RESTORE_AGE_IDENTITY` e `CONFIRM_RESTORE=isolated-target`, e execute `restore.sh ARQUIVO.dump.age` substituindo o entrypoint. Exemplo após configurar as variáveis no shell da estação autorizada:

```sh
docker compose --env-file deploy/.env.production -f compose.production.yaml run --rm \
  --entrypoint /usr/local/bin/restore.sh \
  -e RESTORE_DATABASE_URL -e RESTORE_AGE_IDENTITY -e CONFIRM_RESTORE \
  -v /caminho/identidade.agekey:/run/secrets/restore-key:ro \
  backup /backups/ARQUIVO.dump.age
```

Nesse exemplo `RESTORE_AGE_IDENTITY=/run/secrets/restore-key`. A URL de restauração precisa apontar para o banco separado. O script não usa DROP/CLEAN: conflitos abortam a transação. Valide contagens, contas e pedidos restaurados, e só então planeje recuperação operacional. Rotacione/revogue sessões e credenciais conforme o motivo da recuperação; reenvio SMTP após restauração deve ser decidido antes de iniciar o worker.

## 5. Ativar monitoramento e entrega de alertas

Grave apenas o valor de `METRICS_TOKEN` em `deploy/metrics-token`. O arquivo precisa ser legível pelo UID do Prometheus (65534), mas não público: no host Linux, use proprietário 65534 e modo 0400, com diretório protegido. Não inclua esse arquivo em Git/imagem/logs.

```sh
docker compose --env-file deploy/.env.production -f compose.production.yaml --profile monitoring up -d prometheus
```

Dashboard publicado somente em `127.0.0.1:9090`, acessível por túnel SSH. Regras cobrem indisponibilidade da API, erros 5xx, fila SMTP atrasada/esgotada e reservas expiradas não processadas. Métricas têm autenticação e não incluem e-mails, endereços ou tokens.

`deploy/alertmanager.yml.example` prepara SMTP/destinatário operacional. Escolha um Alertmanager privado/serviço equivalente, preencha sua configuração fora do Git e adicione `alerting.alertmanagers` em Prometheus apontando para ele. O Compose não inicia Alertmanager nem entrega notificações automaticamente. Valide um alerta de teste e sua resolução no destino. Monitore também espaço em disco, recursos do host, expiração TLS e idade dos backups.

## 6. Repetir verificação de release

Pipeline em `.github/workflows/security.yml`: Go race/vet/Postgres, govulncheck, Svelte check/build, npm audit e Trivy HIGH/CRITICAL nas três imagens finais. Ele será executado quando este diretório estiver em um repositório GitHub com Actions habilitado; não houve execução remota de CI nesta etapa.

O smoke completo usa fixtures descartáveis e configuração derivada do Compose de produção. Requer Python 3, OpenSSL, Docker/Compose e imagens construídas nos nomes abaixo:

```sh
docker build -t forma-api api
docker build -t forma-web:release-check web
docker build -t forma-backup:release-check deploy/backup
python3 scripts/deployment_smoke.py
```

O projeto de teste se chama `forma-release-test`; não execute duas instâncias simultaneamente nem reutilize esse nome para dados reais. O script valida HTTPS com CA local confiável pelo cliente de teste, SMTP STARTTLS, Postgres TLS e usuário restrito, assinatura BFF, cookies, contas, checkout, pagamento do adaptador interno, métricas e backup/restauração. Encerra e remove containers/volumes das fixtures. Não envia mensagens externas, não cobra e não altera o banco de desenvolvimento.

## Antes de abrir acesso público

Preencha configurações reais e valide no destino; ative timer/cópia externa/alertas; revise catálogo e atendimento; defina políticas de retenção/exclusão de contas, pedidos, endereços e contatos e os textos comerciais/privacidade. A limpeza técnica não substitui essas decisões. Limites funcionais e futuras integrações estão em [production-readiness.md](production-readiness.md).

Somente os adaptadores de pagamento/envio simulam resultados no servidor. Nenhuma chave bancária/cartão é coletada. A aparência e o fluxo do frontend permanecem de loja normal.
