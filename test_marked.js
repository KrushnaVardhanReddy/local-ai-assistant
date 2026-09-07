const { marked } = require('marked');

async function main() {
    console.log(typeof marked.parseAsync);
    console.log(await marked.parseAsync("# hello"));
}
main();
