package page

const Template = `
<!DOCTYPE html>
<html lang="zh-Hans">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1"> 
    <title>IP Query</title>
    <meta name="description" content="IP Query Tools, open-source project https://github.com/soulteary/ip-helper">
    <link rel="canonical" href="%DOMAIN%%DOCUMENT_PATH%"/>
    <base href="/">
    <style>
      * {
        margin: 0;
        padding: 0;
        box-sizing: border-box;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      }

      body {
        background-color: #f5f5f5;
        padding: 20px;
        display: flex;
        flex-direction: column;
        align-items: center;
        min-height: 100vh;
      }

      .container {
        display: flex;
        min-height: 100vh;
      }

      .sidebar {
        width: 30%;
        padding: 20px;
        display: flex;
        flex-direction: column;
        align-items: center;
      }

      .logo {
        max-width: 200px;
        height: auto;
        border-radius: 16px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        margin-top: 30px;
      }

      .main-content {
        width: 70%;
        padding: 20px;
      }

      .search-container {
        display: flex;
        gap: 10px;
        margin-bottom: 20px;
        width: 100%;
        max-width: 600px;
      }

      .search-input {
        flex: 1;
        padding: 12px 15px;
        font-size: 16px;
        border: 1px solid #ddd;
        border-radius: 4px;
        outline: none;
      }

      .search-button {
        padding: 12px 24px;
        font-size: 16px;
        background-color: #f0f0f0;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        transition: background-color 0.2s;
        border: 1px solid #ddd;
      }

      .search-button:hover {
        background-color: #e0e0e0;
      }

      .result-container {
        background-color: white;
        padding: 20px;
        border-radius: 8px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        width: 100%;
        max-width: 600px;
        margin-bottom: 20px;
      }

      .result-row {
        display: flex;
        margin-bottom: 15px;
        line-height: 1.5;
      }

      .result-label {
        width: 100px;
        color: #666;
      }

      .result-value {
        flex: 1;
        color: #333;
      }

      .command-container {
        background-color: white;
        padding: 20px;
        border-radius: 8px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        width: 100%;
        max-width: 600px;
      }

      .command-title {
        font-size: 16px;
        color: #333;
        margin-bottom: 15px;
      }

      .command-row {
        display: flex;
        margin-bottom: 10px;
      }

      .command-label {
        width: 120px;
        color: #666;
      }

      .command-code {
        flex: 1;
        color: #333;
        font-family: monospace;
      }

      .result-label,.command-title{
        user-select: none;
      }

      .logo:hover {
        opacity: 0.8;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div class="sidebar">
        <h1>IP QUERY</h1>
        <img src="panda.jpg" alt="Logo" class="logo" width="200" height="357"/>
      </div>
      <div class="main-content">
        <div class="search-container">
          <form action="/" method="post">
            <input type="text" name="ip" class="search-input" placeholder="请输入要查询的 IP 地址" value="%IP_ADDR%" />
            <button class="search-button" type="submit">查询</button>  
          </form>
        </div>

        <div class="result-container">
          <div class="result-row">
            <div class="result-label">IP</div>
            <div class="result-value">%IP_ADDR%</div>
          </div>
          <div class="result-row">
            <div class="result-label">地址</div>
            <div class="result-value">%DATA_1_INFO%</div>
          </div>
          <!-- <div class="result-row">
            <div class="result-label">运营商</div>
            <div class="result-value">联通</div>
          </div> -->
          <div class="result-row">
            <div class="result-label">数据二</div>
            <div class="result-value">等待接入</div>
          </div>
          <div class="result-row">
            <div class="result-label">数据三</div>
            <div class="result-value">等待接入</div>
          </div>
          <div class="result-row">
            <div class="result-label">URL</div>
            <div class="result-value">%DOMAIN%/%IP_ADDR%</div>
          </div>
        </div>

        <div class="command-container">
          <div class="command-title">命令行查询详细信息</div>
          <div class="command-row">
            <div class="command-label">
              <label for="command-curl">UNIX/Linux</label>
            </div>
            <div class="command-code">
              <span class="command-prompt">#</span>
              <input id="command-curl" type="text" value="curl %ONLY_DOMAIN_WITH_PORT%" />
            </div>
          </div>
          <div class="command-row">
            <div class="command-label">
              <label for="command-telenet">Windows</label>
            </div>
            <div class="command-code">
              <span class="command-prompt">></span>
              <input id="command-telenet" type="text" value="telnet %ONLY_DOMAIN%" />
            </div>
          </div>
          <div class="command-row">
            <div class="command-label">
              <label for="command-ftp">其他</label>
            </div>
            <div class="command-code">
              <span class="command-prompt">></span>
              <input id="command-ftp" type="text" value="ftp %ONLY_DOMAIN%" />
            </div>
          </div>

          <div class="command-title" style="margin-top: 20px">仅查询IP</div>
          <div class="command-row">
            <div class="command-label">
              <label for="command-curl-ip-only">UNIX/Linux</label>
            </div>
            <div class="command-code">
              <span class="command-prompt">#</span>
              <input id="command-curl-ip-only" type="text" value="curl %ONLY_DOMAIN_WITH_PORT%/ip" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </body>
</html>
`
