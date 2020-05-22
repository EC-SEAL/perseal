const proxy = [
{
context: '/api',
target: 'http://localhost:8082',
pathRewrite: {'^/api': ''}
}
];
module.exports = proxy;
