// Proxy do ng serve (apenas desenvolvimento).
//
// Cada microservico tem o seu prefixo, e o destino vem do ambiente porque ele
// muda conforme onde o dev server roda:
//   - dentro do compose: http://inventory:8000  (nome do servico na rede)
//   - direto no host:    http://localhost:8000 (porta publicada)
// Dentro do container, localhost seria o proprio frontend.
module.exports = {
  '/api/inventory': {
    target: process.env['INVENTORY_API_URL'] || 'http://localhost:8000',
    secure: false,
    changeOrigin: true,
    pathRewrite: { '^/api/inventory': '' }
  },
  '/api/billing': {
    target: process.env['BILLING_API_URL'] || 'http://localhost:8001',
    secure: false,
    changeOrigin: true,
    pathRewrite: { '^/api/billing': '' }
  }
};
