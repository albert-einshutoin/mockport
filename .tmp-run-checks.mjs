const { execSync } = require('child_process');
const cwd = '/Volumes/Satechi/Developer/mockport';

function run(cmd) {
  console.log(`===== COMMAND: ${cmd} =====`);
  try {
    const out = execSync(cmd, { cwd, encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'] });
    if (out) process.stdout.write(out);
    console.log('EXIT_CODE=0');
    return 0;
  } catch (e) {
    if (e.stdout) process.stdout.write(e.stdout);
    if (e.stderr) process.stderr.write(e.stderr);
    console.log(`EXIT_CODE=${e.status ?? 1}`);
    return e.status ?? 1;
  } finally {
    console.log();
  }
}

run('bash scripts/check-doc-links.sh');
run('bash scripts/check-public-trust.sh');
