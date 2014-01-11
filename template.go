package main

var buildPage string = `<html>
                       <body>
                       <form action='/build' method='post'>
                       repo url:
                       <input type="text" name="repo">
                       <br />
                       rev(optional):
                       <input type="text" name="rev">
                       <br />
                       ppa name:
                       <input type="text" name="ppa">
                       <br />
                       <input type="submit" value="submit">
                       </form>
                       </body>
                       </html>`
