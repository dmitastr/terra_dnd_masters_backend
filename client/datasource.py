import logging
from typing import Any

import arrow
import requests
from pydantic import BaseModel


logging.basicConfig(level=logging.INFO,
                    format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)


class Demand(BaseModel):
    vk_id: int
    slots: list[dict[str, Any]]
    players_count: int = 1
    vk_username: str = "NONE"
    first_name: str = "Игрок"
    last_name: str = ""
    for_week: str = ""
    comment: str = ""


class Datasource:
    def __init__(self, access_token: str, api_base: str) -> None:
        self.client = requests.Session()
        self.client.headers.update({"X-Api-Key": f"{access_token}"})
        self.access_token = access_token
        self.api_base = api_base
        self.max_tries = 3

    def get_slots(self, user_id: int, for_week: str = "") -> list[dict[str, Any]]:
        resp = self.client.get(f"{self.api_base}/slots", params={
            "from": arrow.now().date().isoformat(),
            "to": arrow.now().shift(days=7).date().isoformat(),
            "for_week": for_week
        })
        resp.raise_for_status()
        data = resp.json()
        return data["slots"]

    def get_demands(self, user_id: int, for_week: str = "") -> list[Demand]:
        resp = self.client.get(
            f"{self.api_base}/demands/{user_id}", params={"week": for_week})
        resp.raise_for_status()
        data = resp.json()
        return [Demand.model_validate(item) for item in data["demands"]]

    def get_all_demands(self, for_week: str = "") -> list[Demand]:
        resp = self.client.get(
            f"{self.api_base}/demands", params={"week": for_week})
        resp.raise_for_status()
        data = resp.json()
        return [Demand.model_validate(item) for item in data["demands"]]

    def post_demands(self, demand: Demand) -> requests.Response | None:
        payload = {"demands": [demand.model_dump()]}
        log.info(f"Sending demand: {payload}")
        resp = self.client.post(f"{self.api_base}/demands", json=payload)

        if resp:
            resp.raise_for_status()
        return resp

    def delete_demands(self, user_id: int, for_week: str = "") -> requests.Response | None:
        log.info(f"Deleting demands for user {user_id} for week {for_week}")
        resp = self.client.delete(
            f"{self.api_base}/demands/{user_id}", params={"week": for_week})
        if resp:
            resp.raise_for_status()
        return resp
