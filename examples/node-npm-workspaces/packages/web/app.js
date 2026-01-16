import { APP_NAME, formatMessage } from "@monorepo/shared";

document.getElementById("app-name").textContent = APP_NAME;
console.log(formatMessage(`${APP_NAME} frontend loaded`));

document.getElementById("content").innerHTML = `
  <p>Welcome to the ${APP_NAME} web application!</p>
  <p>This demonstrates a full-stack monorepo with shared code.</p>
`;
