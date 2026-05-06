from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .routers.runs import router as runs_router
from .routers.workflows import router as workflows_router

app = FastAPI(title="MiniFlow API")
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])
app.include_router(workflows_router)
app.include_router(runs_router)


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok"}
