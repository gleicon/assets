Construir um programa health checker concorrente que:
Monitora múltiplos serviços simultaneamente
Implementa timeouts e lógica de retry
Agrega resultados de múltiplos endpoints
Reportar status em tempo real

Funcionalidades:
Configuração de serviços a serem monitorados (dica: usar o config file feito anteriormente usando cobra-cli)
Health checks concorrentes para 10+ serviços http e https
Lógica configurável de timeout e retry
Output de relatório de status em JSON
