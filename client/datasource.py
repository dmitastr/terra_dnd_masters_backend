import logging
from typing import Any

import arrow
import requests

logging.basicConfig(level=logging.INFO,
                    format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)


class Datasource:
    def __init__(self, access_token: str, api_base: str) -> None:
        self.client = requests.Session()
        self.client.headers.update({"X-Api-Key": f"{access_token}"})
        self.access_token = access_token
        self.api_base = api_base
        self.max_tries = 3

    def get_slots(self, user_id: int) -> list[dict[str, Any]]:
        resp = self.client.get(f"{self.api_base}/slots", params={
            "from": arrow.now().date().isoformat(),
            "to": arrow.now().shift(days=7).date().isoformat()
        })
        resp.raise_for_status()
        data = resp.json()
        return data["slots"]

    def get_demands(self, user_id: int, for_week: str = "") -> list[dict[str, Any]]:
        resp = self.client.get(
            f"{self.api_base}/demands/{user_id}", params={"week": for_week})
        resp.raise_for_status()
        data = resp.json()
        return data["demands"]

    def get_all_demands(self, for_week: str = "") -> list[dict[str, Any]]:
        resp = self.client.get(
            f"{self.api_base}/demands", params={"week": for_week})
        resp.raise_for_status()
        data = resp.json()
        return data["demands"]

    def post_demands(self, vk_id: int, slots: list[dict[str, Any]], players_count: int = 1, first_name: str = "Игрок", last_name: str = "", for_week: str = "", username: str = "NONE") -> requests.Response | None:
        payload = {"demands": [
            {"vk_id": vk_id,
             "slots": slots,
             "players_count": players_count,
             "vk_username": username,
             "first_name": first_name,
             "last_name": last_name,
             "for_week": for_week}]}
        log.info(f"Sending demand: {payload}")
        resp = self.client.post(f"{self.api_base}/demands", json=payload)

        if resp:
            resp.raise_for_status()
        return resp
