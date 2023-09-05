const puppeteer = require("puppeteer");
let browser
(async () => {
  const html = await fetch('https://brutelogic.com.br/xss.php?c1=%3C/script%3E%3Csvg%3E%3Cscript%3Ealert(1)-%26apos%3B').then(res => res.text());

  browser = await puppeteer.launch({
      headless: true,
      ignoreHTTPSErrors: true,
      args: [
        "--proxy-server='direct://'",
        '--proxy-bypass-list=*',
        '--disable-gpu',
        '--disable-dev-shm-usage',
        '--disable-setuid-sandbox',
        '--no-first-run',
        '--no-sandbox',
        '--no-zygote',
        '--single-process',
        '--ignore-certificate-errors',
        '--ignore-certificate-errors-spki-list',
        '--enable-features=NetworkService'
      ]
    });
  const [page] = await browser.pages();

  const dialogDismissed = new Promise((resolve, reject) => {
    let wait = setTimeout(() => {
      clearTimeout(wait);
      process.exit(); //Timed out after 4s..
    }, 10000)

    const handler = async dialog => {
      await dialog.dismiss();
      resolve(dialog.message());
    };

    page.once("dialog", handler);
  });

  await page.setContent(html);
  const msg = await dialogDismissed;
  console.log("[Alerta] "+msg); // => hello world
  await page.close();
})()
  .catch(err => console.error(err))
  .finally(() => browser?.close())
;
