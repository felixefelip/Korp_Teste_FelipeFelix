// Proxy do ng serve (apenas desenvolvimento).
//
// O destino vem do ambiente porque ele muda conforme onde o dev server roda:
//   - dentro do compose: http://estoque:8000  (nome do servico na rede)
//   - direto no host:    http://localhost:8000 (porta publicada)
// Dentro do container, localhost seria o proprio frontend.
module.exports = {
  '/api': {
    target: process.env['API_URL'] || 'http://localhost:8000',
    secure: false,
    changeOrigin: true,
    pathRewrite: { '^/api': '' }
  }
};
