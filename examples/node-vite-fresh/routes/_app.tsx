import { app } from "./$_app.ts";

export default app(function App(props) {
  return (
    <html lang="en">
      <head>
        <meta charset="UTF-8" />
        <title>Fresh on Railpack</title>
      </head>
      <body>{props.children}</body>
    </html>
  );
});
