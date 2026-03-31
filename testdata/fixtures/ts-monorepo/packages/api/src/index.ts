import express from "express";
import { formatResponse } from "@ts-monorepo/shared";

const app = express();
app.use(express.json());

app.get("/health", (_req, res) => {
  res.json(formatResponse({ status: "ok" }));
});

app.get("/users/:id", (req, res) => {
  const { id } = req.params;
  res.json(formatResponse({ id, name: `User ${id}` }));
});

const port = process.env.PORT ?? 3000;
app.listen(port, () => {
  console.log(`API listening on port ${port}`);
});

export { app };
