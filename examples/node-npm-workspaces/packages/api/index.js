import express from "express";
import { APP_NAME, API_VERSION, formatMessage, createResponse } from "@monorepo/shared";

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get("/", (req, res) => {
  console.log(formatMessage("Health check requested"));
  res.json(createResponse({
    app: APP_NAME,
    version: API_VERSION,
    message: "API is running"
  }));
});

app.get("/api/users", (req, res) => {
  console.log(formatMessage("Users list requested"));
  res.json(createResponse({
    users: [
      { id: 1, name: "Alice" },
      { id: 2, name: "Bob" }
    ]
  }));
});

app.listen(PORT, () => {
  console.log(formatMessage(`${APP_NAME} API v${API_VERSION} listening on port ${PORT}`));
});
