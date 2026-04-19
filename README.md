# certcheck

Monitore a expiração de certificados SSL e o vencimento de registros de domínio diretamente pelo terminal — uma alternativa gratuita a serviços pagos como o Datadog.

## Instalação

### Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)

### Compilando a partir do código fonte

```bash
git clone https://github.com/NSantos6/certcheck.git
cd certcheck
go build -o certcheck .
```

Mova o binário para um diretório no seu `$PATH` para usar de qualquer lugar:

```bash
mv certcheck /usr/local/bin/
```

## Comandos

### `ssl` — Verificar certificado SSL

Verifica a data de expiração do certificado HTTPS de um ou mais domínios.

```bash
# Um domínio
certcheck ssl exemplo.com.br

# Múltiplos domínios
certcheck ssl exemplo.com.br google.com github.com

# A partir de um arquivo
certcheck ssl --file dominios.yaml

# Alertar quando faltar menos de 14 dias (padrão: 30)
certcheck ssl exemplo.com.br --warn-days 14

# Saída em JSON
certcheck ssl exemplo.com.br --json
```

### `domain` — Verificar registro de domínio

Consulta o WHOIS do domínio para verificar o vencimento do registro. Suporta domínios `.br` via Registro.br.

```bash
# Um domínio
certcheck domain meusite.com.br

# Múltiplos domínios
certcheck domain meusite.com.br outrosite.net.br

# A partir de um arquivo
certcheck domain --file dominios.yaml

# Alertar quando faltar menos de 90 dias (padrão: 60)
certcheck domain meusite.com.br --warn-days 90

# Saída em JSON
certcheck domain meusite.com.br --json
```

### `scan` — Verificação completa

Verifica SSL e registro de domínio ao mesmo tempo, em paralelo.

```bash
# Um domínio
certcheck scan meusite.com.br

# A partir de um arquivo
certcheck scan --file dominios.yaml

# Thresholds customizados
certcheck scan meusite.com.br --ssl-warn-days 14 --domain-warn-days 90

# Saída em JSON (útil para scripts e CI/CD)
certcheck scan --file dominios.yaml --json

# Enviar alerta por email se houver domínios expirando
certcheck scan --file dominios.yaml --notify eu@email.com --smtp-user eu@gmail.com
```

## Arquivo de domínios

Você pode criar um arquivo YAML com a lista de domínios para não precisar digitá-los toda vez:

```yaml
# dominios.yaml
- meusite.com.br
- outrosite.net.br
- google.com
- github.com
```

## Exemplos de saída

### Tabela (padrão)

```
SSL Certificates
  DOMAIN          | EXPIRES    | DAYS LEFT | ISSUER                | STATUS
------------------+------------+-----------+-----------------------+--------------
  meusite.com.br  | 2025-08-10 |        82 | Let's Encrypt         | OK
  outrosite.net   | 2025-05-01 |         8 | Sectigo Limited       | EXPIRING SOON
  antigossite.com | 2024-12-01 |       -50 | -                     | EXPIRED

Domain Registration
  DOMAIN          | EXPIRES    | DAYS LEFT | REGISTRAR             | STATUS
------------------+------------+-----------+-----------------------+---------
  meusite.com.br  | 2027-03-15 |       695 | Universo Online S.A.  | OK
  outrosite.net   | 2025-06-01 |        39 | -                     | EXPIRING SOON
```

### JSON (`--json`)

```json
[
  {
    "domain": "meusite.com.br",
    "ssl": {
      "expires_at": "2025-08-10",
      "days_left": 82,
      "issuer": "Let's Encrypt"
    },
    "registration": {
      "expires_at": "2027-03-15",
      "days_left": 695
    }
  }
]
```

## Notificações por email

O comando `scan` pode enviar um email de alerta quando algum domínio estiver expirando ou já expirado. O envio é **opcional** — só acontece se o flag `--notify` for informado.

O email só é enviado se houver domínios dentro do threshold configurado (`--ssl-warn-days` / `--domain-warn-days`). Se tudo estiver OK, nenhum email é enviado.

```bash
certcheck scan --file dominios.yaml \
  --notify destino@email.com \
  --smtp-host smtp.gmail.com \
  --smtp-port 587 \
  --smtp-user remetente@gmail.com \
  --smtp-pass suasenha
```

### Flags de email

| Flag | Padrão | Descrição |
|---|---|---|
| `--notify` | — | Email de destino (obrigatório para ativar) |
| `--smtp-host` | `smtp.gmail.com` | Servidor SMTP |
| `--smtp-port` | `587` | Porta SMTP |
| `--smtp-user` | — | Usuário SMTP |
| `--smtp-pass` | — | Senha SMTP |
| `--smtp-from` | smtp-user | Endereço de envio |

### Usando variável de ambiente (recomendado)

Para evitar expor a senha no histórico do terminal, use a variável de ambiente `CERTCHECK_SMTP_PASS`:

```bash
export CERTCHECK_SMTP_PASS=suasenha
certcheck scan --file dominios.yaml --notify eu@email.com --smtp-user eu@gmail.com
```

### Usando com Gmail

No Gmail, você precisa usar uma **senha de app** (não a senha da conta). Gere uma em: `Conta Google → Segurança → Verificação em duas etapas → Senhas de app`.

## Uso em CI/CD

O `certcheck` retorna exit code `1` quando ocorre um erro, facilitando o uso em pipelines:

```yaml
# Exemplo: GitHub Actions
- name: Verificar certificados
  run: ./certcheck scan --file dominios.yaml --json
```

## Cores e status

| Status | Significado |
|---|---|
| `OK` | Dentro do prazo |
| `EXPIRING SOON` | Expira dentro do threshold configurado |
| `EXPIRED` | Já expirado |
| `ERROR` | Não foi possível verificar |

## Suporte a domínios .br

O `certcheck` consulta diretamente o WHOIS do [Registro.br](https://registro.br) para domínios `.br`, `.com.br`, `.net.br`, `.org.br`, entre outros. Não é necessária nenhuma configuração adicional.

## Deploy no OpenShift / Kubernetes

### Estrutura

```
k8s/
├── configmap.yaml   # lista de domínios
├── secret.yaml      # credenciais SMTP
└── cronjob.yaml     # CronJob (padrão: todo dia às 8h)
```

### 1. Build e push da imagem

```bash
docker build -t registry.interno.exemplo.com/certcheck:latest .
docker push registry.interno.exemplo.com/certcheck:latest
```

> Atualize o campo `image:` no `cronjob.yaml` com o endereço do seu registry interno.

### 2. Editar os arquivos

**`k8s/configmap.yaml`** — adicione seus domínios:
```yaml
data:
  domains.yaml: |
    - meusite.com.br
    - outrosite.net.br
```

**`k8s/secret.yaml`** — preencha as credenciais SMTP:
```yaml
stringData:
  smtp-user: "remetente@gmail.com"
  smtp-pass: "sua-senha-de-app"
  notify-to: "destino@email.com"
```

> Nunca commite o `secret.yaml` com dados reais no Git.

### 3. Aplicar no cluster

```bash
oc new-project certcheck          # ou: kubectl create namespace certcheck
oc apply -f k8s/configmap.yaml
oc apply -f k8s/secret.yaml
oc apply -f k8s/cronjob.yaml
```

### 4. Testar manualmente

```bash
# Dispara o job imediatamente sem esperar o schedule
oc create job certcheck-teste --from=cronjob/certcheck

# Acompanha os logs
oc logs -l job-name=certcheck-teste -f
```

### Ajustar o schedule

O campo `schedule` no `cronjob.yaml` usa sintaxe cron padrão:

```yaml
schedule: "0 8 * * *"    # todo dia às 8h
schedule: "0 8 * * 1"    # toda segunda às 8h
schedule: "0 8,20 * * *" # duas vezes por dia: 8h e 20h
```

### Compatibilidade com OpenShift

O Dockerfile e o CronJob já estão configurados para rodar como usuário não-root (`UID 1001`), com `readOnlyRootFilesystem` e sem capabilities — compatível com o SCC `restricted` do OpenShift por padrão.

## Licença

MIT
