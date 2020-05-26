const proxy = [
{
context: '/api',
target: host,
pathRewrite: {'^/api': ''}
}
];
module.exports = proxy;
