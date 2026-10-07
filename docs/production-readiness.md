# Preparação para produção

A loja usa contas, catálogo, estoque, cotações e pedidos persistidos no Postgres. Somente pagamento/envio usam adaptadores simulados no servidor; o frontend mantém o fluxo normal de compra.

## Base técnica implementada

- Senhas Argon2id; sessões com hash no banco e cookies HttpOnly/Secure/__Host; confirmação de e-mail, recuperação de senha e gestão/revogação de sessões.
- Registro, desafio e e-mail em transação; outbox criptografada, SMTP TLS e retries. Cadastro/recuperação não revelam se um e-mail existe.
- Preços e reserva de estoque transacionais; cotações vinculadas ao cliente, CEP e itens; idempotência; autorização por proprietário; vencimento e restituição de reservas.
- CSP com nonce, proteção CSRF, JSON limitado por tamanho/tempo, quotas compartilhadas no Postgres e IP encaminhado assinado pelo BFF.
- Compose de produção, HTTPS Caddy, adapter Node, API privada, credencial restrita de runtime e migrations separadas.
- Métricas privadas, regras de alerta, backup age criptografado e scripts de restauração; pipeline de testes e scans de imagens.

Testes Go com race detector/Postgres, vet, check/build/audit npm e govulncheck passaram. Trivy não detectou vulnerabilidades altas/críticas nas imagens API/web/backup em 07/10/2026. O smoke isolado de produção validou TLS, SMTP, ciclo da conta, compra e restauração. Evidências: [revisão de segurança](security-review.md).

## Configurações que dependem do destino

[Guia de publicação](publication.md): escolher hospedagem/domínio, Postgres TLS e SMTP; preencher segredos; apontar DNS; executar migrations; ativar backup agendado/cópia externa e entrega de alertas; repetir smoke no destino autorizado. Nada foi publicado em um host público nesta etapa.

## Limites funcionais e operacionais

- Histórico retorna os 100 pedidos mais recentes; o catálogo é carregado no layout. Para maiores volumes, implementar paginação/filtros de catálogo e histórico no servidor por página.
- Não existe painel administrativo para catálogo, estoque e atendimento. Até sua implementação, operadores precisam de procedimentos controlados e credenciais próprias; não usar o usuário runtime ou expor Postgres. Acesso administrativo futuro deve incluir MFA e autorização específica.
- Mensagens de contato são persistidas, mas não há notificação ao atendimento. Definir a rotina de leitura/atendimento antes de aceitar solicitações públicas.
- Revisar dados, preços, estoque seed e direitos das fotos; preparar armazenamento durável dos assets e textos comerciais/privacidade/termos adequados à operação.
- A limpeza técnica cobre sessões/cotações/desafios expirados, outbox após sete dias e eventos de segurança após 90 dias. Retenção/exclusão de contas, endereços, pedidos e contatos não foi definida nem automatizada: requer política própria.
- Para transportadora real: origem, peso/dimensões, contratação, etiquetas, rastreamento e eventos. O contrato atual cobre cotação.
- Para pagamento real: checkout do provedor, idempotência/referência externa, webhooks autenticados, estados assíncronos, conciliação e reembolsos. O contrato atual é síncrono.

A base de deploy e segurança está preparada e validada localmente; configuração operacional e limites acima devem orientar a publicação, sem confundir configuração pronta com operação pública já ativada.
