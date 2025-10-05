import logging
import os
from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel
import httpx

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[logging.StreamHandler(), logging.FileHandler("app.log")]
)
logger = logging.getLogger(__name__)

app = FastAPI(title="Sensor Manager API")

# Базовый URL smart_home (без пути)
SMART_HOME_BASE = os.getenv("SMART_HOME_AP", "http://smarthome-app:8080")
SMART_HOME_API = f"{SMART_HOME_BASE}/api/v1/sensors"


class SensorCreate(BaseModel):
    name: str
    type: str
    location: str
    unit: str


@app.middleware("http")
async def log_requests(request: Request, call_next):
    logger.info(f"Incoming request: {request.method} {request.url}")
    body = await request.body()
    if body:
        logger.info(f"Request body: {body.decode(errors='ignore')}")
    response = await call_next(request)
    logger.info(f"Response status: {response.status_code}")
    return response


@app.get("/health")
async def health():
    """Healthcheck endpoint."""
    return {"status": "ok"}


@app.post("/api/v1/sensors")
async def create_sensor(sensor: SensorCreate):
    logger.info(f"Forwarding create_sensor request: {sensor.dict()}")
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(SMART_HOME_API, json=sensor.dict())
            logger.info(f"smart_home response [{response.status_code}]: {response.text}")
            response.raise_for_status()
        except httpx.HTTPError as e:
            logger.error(f"Ошибка при вызове smart_home: {e}")
            raise HTTPException(status_code=500, detail=f"Ошибка при вызове smart_home: {str(e)}")

    return response.json()


@app.delete("/api/v1/sensors/{sensor_id}")
async def delete_sensor(sensor_id: str):
    logger.info(f"Forwarding delete_sensor request: id={sensor_id}")
    async with httpx.AsyncClient() as client:
        try:
            response = await client.delete(f"{SMART_HOME_API}/{sensor_id}")
            logger.info(f"smart_home response [{response.status_code}]: {response.text}")
            response.raise_for_status()
        except httpx.HTTPError as e:
            logger.error(f"Ошибка при вызове smart_home: {e}")
            raise HTTPException(status_code=500, detail=f"Ошибка при вызове smart_home: {str(e)}")

    return {"status": "deleted", "id": sensor_id}
