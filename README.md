# greenbook

bupt web课设


```mermaid
graph TD
  A[API Root] --> B(Public APIs)
  A --> C(Protected APIs)
  B --> D[GET /articles]
  B --> E[GET /articles/:id]
  C --> F[POST/PUT/DELETE /articles]
  C --> G[评论操作]
```
/opt/homebrew/opt/openjdk/bin:/opt/homebrew/opt/go/libexec/bin:/opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:/System/Cryptexes/App/usr/bin:/usr/bin:/bin:/usr/sbin:/sbin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/local/bin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/bin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/appleinternal/bin:/opt/homebrew/opt/python@3.13