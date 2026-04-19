# certcheck

Monitore a expiração de certificados SSL e o vencimento de registros de domínio diretamente pelo terminal — uma alternativa gratuita a serviços pagos como o Datadog.

## Instalação

### Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)

### Compilando a partir do código fonte

```bash
git clone https://github.com/nmesquitasantos/certcheck.git
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

## Licença

MIT
