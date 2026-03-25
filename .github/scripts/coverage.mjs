#!/usr/bin/env zx

const expected = 100

cd(path.resolve(__dirname, '..', '..'))

await spinner('Running tests', async () => {
  await $`go test -coverprofile=coverage.out -coverpkg=github.com/maml-dev/go-maml/... ./...`
  await $`go tool cover -html=coverage.out -o coverage.html`
})

const cover = await $({verbose: true})`go tool cover -func=coverage.out`
const total = +cover.stdout.match(/total:\s+\(statements\)\s+(\d+\.\d+)%/)[1]
if (total < expected) {
  echo(chalk.red(`Coverage is too low: ${total}% < ${expected}% (expected)`))
  process.exit(1)
} else {
  echo(`Coverage is good: ${chalk.green(total + '%')} >= ${expected}% (expected)`)
}
