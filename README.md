# ledger-exp

## Estrutura de tabelas
![img.png](.github/images/img.png)

## Arquitetura da aplicação
![architecture.png](.github/images/architecture.png)

- Nestre projeto foi utilizado o [PostgreSQL](https://www.postgresql.org/) como solução de banco de dados, o [Redis](https://redis.io/) foi escolhido para camada de cache 
e controle de concorrencia([distlock](https://redis.io/docs/reference/patterns/distributed-locks/)).
- Na parte de observabilidade a escolha foi pelo projeto open source [OpenTelemetry](https://opentelemetry.io/), que possibilita a distribuição de métricas e spans
para diferentes provedores de forma agnóstica.

### Pacotes
![img.png](./.github/images/packages.png)
 - A pasta /cmd contém o bootstrap/main da aplicação.
 - Dentro de /internal foram distribuidos os pacotes com o contexto de négocio.
   1. **accounts** -> Gestão das contas dos postadores;
   2. **api** -> Implementação dos handlers http;
   3. **balances** -> Gestão dos saldos das contas, disponibilizado pela view **transactions_balances**;
   4. **holders** -> Gestão dos postadores;
   5. **statements** -> Apresentação do extrato da conta, baseado nas transações. Feature separada do package **transactions** para prover maior autonomia de filtros;
   6. **transactions** -> Gestão das transações realizadas, como créditos, débitos e transferências entre contas;
 - Em /migrations disponibilizado todos os scripts sql (DDL) para migração do banco de dados.
 - Em /pkg estão disponíveis todos pacotes utilizados para criação da aplicação, estes que não possuem relação com o negócio.

## Como rodar o projeto?
```shell
docker-compose --env-file .env-docker  up -d
```
Após a inicialização de todos os serviços, eles estaram expostos em:
_A migração do banco de dados irá ser executada juntamente com a inicialição do docker-compose_
- ledger-exp-api -> [localhost:8080](http://localhost:8080)
- postgres -> [localhost:5432](postgres://ledger-exp:ledger-exp@localhost:5432/ledger-exp?sslmode=disable)
- redis-commander -> [localhost:8081](http://localhost:8081/) root - root
- jaeger -> [localhost:16686](http://localhost:16686/search)
- prometheus -> [localhost:9090](http://localhost:9090/graph)

## Como rodar os testes unitários e de integração?
Unitários:
```shell
go test -json -race -tags=unit ./...
```

Integração:
```shell
go test -json -race -tags=integration ./...
```

## Como testar?
Sugiro utilizar o arquivo exportado do Insomnia, já possui todas os endpoints mapeados com exemplos de utilização.
[Arquivo](./Insomnia_ledger-exp.json)

Porém para um fluxo consistente, devemos chamar os seguintes endpoints:
1. POST /v1/holders
2. POST /v1/accounts
3. Criar transações
   1. POST /v1/transactions/credits -> realiza um crédito na conta.
   2. POST /v1/transactions/debits -> realiza um débito na conta.
   3. POST /v1/transactions/p2p -> realiza uma transferência entre contas.
4. GET /v1/accounts/:accountID/statements -> extrato da conta.
5. GET /v1/accounts/:accountID/balances -> consulta saldo da conta.

## Curiosidades
1. Como funciona a geração dos mocks utilizados nos testes?
   - Na pasta ./scripts existe um arquivo shell onde todos os arquivos/interfaces são mapeados para gerar um mock.
2. Como a migração do banco de dados acontece? 
   - A migração do banco de dados é feita pelo service database-migration, definido no [docker-compose.yml](./docker-compose.yml).
3. Como a observabilidade funciona?
   - Foi utilizado o service OpenTelemetry Collector definido no [docker-compose.yml](./docker-compose.yml), para recepcionar todos os spans e métricas geradas pela aplicação e os trasmite ao Jaeger e Prometheus, respectivamente.
