"""
VK-бот: сбор заявки (количество игроков + выбор временных слотов).

ЛОГИКА ДИАЛОГА
--------------
Шаг 1 (players) — пользователь жмёт "−"/"+", чтобы выбрать количество игроков
                  (минимум 1), кнопка с числом ничего не делает, "Далее ▶"
                  переключает на шаг 2.
Шаг 2 (slots)   — показывается до MAX_SLOTS кнопок-слотов на "странице" +
                  кнопки "◀"/"▶" для переключения страниц и "Отправить ✅".
                  Клик по слоту переключает его выделение (зелёный <-> белый)
                  и добавляет/убирает индекс слота в slots_selected.
                  "Отправить" завершает диалог итоговым сообщением.

Список слотов больше не статический: при каждом "/start" бот получает его
через ваш класс Datasource (метод get_shifts) и сохраняет в UserState.slots —
у каждого пользователя может быть свой набор смен, и при повторном "/start"
список запрашивается заново.

ВАЖНОЕ ПРЕДПОЛОЖЕНИЕ: код ожидает, что Datasource.get_shifts(user_id)
возвращает список объектов "смена", у каждого из которых есть поля id и title
(человекочитаемое название слота). Если в вашем классе поля называются иначе
(например name вместо title, или это dict, а не объект) — поправьте только
две функции shift_id() и shift_label() ниже, остальной код трогать не нужно.

РЕАЛИЗАЦИЯ
----------
Все кнопки — callback-кнопки (type="callback"). Нажатие на такую кнопку НЕ
отправляет сообщение в чат, а генерирует событие `message_event`, которое
бот получает через Bot Long Poll API и обрабатывает, после чего перерисовывает
(messages.edit) то же самое сообщение с новой клавиатурой. Поэтому у каждого
пользователя в чате остаётся одно "живое" сообщение на каждом шаге, а не лог
из десятков сообщений.

НАСТРОЙКА ПЕРЕД ЗАПУСКОМ
------------------------
1. pip install vk_api
2. В настройках сообщества (Управление -> Работа с API -> Long Poll API):
   - включить Long Poll API;
   - версия API: Bot API (не Callback API сервера);
   - в "Типы событий" включить как минимум: "Входящие сообщения" (message_new)
     и "Нажатие на кнопку" (message_event).
3. Получить токен сообщества с правом messages (Управление -> Работа с API ->
   Ключи доступа) и заполнить GROUP_TOKEN, GROUP_ID ниже.
4. Поправить импорт `from datasource import Datasource` под реальный путь
   вашего модуля и передать в Datasource(...) то, что требует его конструктор
   (подключение к БД, конфиг и т.п.) — в блоке `if __name__ == "__main__"`.
"""

import json
import logging
from dataclasses import dataclass, field
from enum import Enum
import os
from typing import Any, Dict, List, Optional, Set

import vk_api
from vk_api.bot_longpoll import VkBotEventType, VkBotLongPoll
from vk_api.keyboard import VkKeyboard, VkKeyboardColor

# ваш класс — поправьте путь импорта при необходимости
from datasource import Datasource

# ─────────────────────────────── НАСТРОЙКИ ───────────────────────────────

# токен сообщества (Bot API, права: messages)
TOKEN = os.environ["VK_TOKEN"]
GROUP_ID = int(os.environ["VK_GROUP_ID"])
API_BASE = os.environ["API_BASE"].rstrip("/")  # e.g. http://localhost:8000
API_TOKEN = os.environ["CLIENT_API_KEY"]           # числовой id сообщества

MAX_SLOTS = 5      # сколько кнопок-слотов показывается на одной странице
MIN_PLAYERS = 1     # минимальное количество игроков

logging.basicConfig(level=logging.INFO,
                    format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)


def shift_id(shift: Any) -> Any:
    """
    Достаёт идентификатор смены — он используется как ключ в payload кнопки
    и в slots_selected. Если у объекта, который возвращает ваш Datasource,
    поле называется иначе (например shift_id вместо id) — поправьте только
    эту строку.
    """
    return shift.id


def shift_label(shift: Any) -> str:
    """
    Достаёт человекочитаемое название смены для подписи на кнопке
    (например "Пн 18:00–20:00"). Поправьте под реальное поле вашего Shift —
    например shift.name или shift["title"], если это dict.
    """
    return shift.title


# ─────────────────────────────── СОСТОЯНИЕ ───────────────────────────────

class Step(str, Enum):
    PLAYERS = "players"
    SLOTS = "slots"
    DONE = "done"


@dataclass
class UserState:
    """Состояние диалога одного пользователя."""

    step: Step = Step.PLAYERS
    players: int = MIN_PLAYERS
    # объекты смен из Datasource.get_shifts
    slots: List[Any] = field(default_factory=list)
    slot_page: int = 0
    # id выбранных смен (shift_id(...))
    slots_selected: Set[Any] = field(default_factory=set)
    peer_id: int = 0
    message_id: int = 0  # id сообщения с клавиатурой, которое мы перерисовываем


# user_id -> UserState. Для прод-варианта стоит заменить на Redis/БД,
# чтобы состояние переживало перезапуск бота.
sessions: Dict[int, UserState] = {}


def get_state(user_id: int) -> UserState:
    if user_id not in sessions:
        sessions[user_id] = UserState()
    return sessions[user_id]


def reset_state(user_id: int) -> UserState:
    sessions[user_id] = UserState()
    return sessions[user_id]


# ─────────────────────────────── КЛАВИАТУРЫ ───────────────────────────────

def _payload(**kwargs) -> dict:
    return kwargs


def build_players_keyboard(players: int) -> VkKeyboard:
    """Клавиатура шага 1: выбор количества игроков."""
    kb = VkKeyboard(inline=True)

    kb.add_callback_button("−", color=VkKeyboardColor.NEGATIVE,
                           payload=_payload(action="players_minus"))
    kb.add_callback_button(f"Игроков: {players}", color=VkKeyboardColor.SECONDARY,
                           payload=_payload(action="noop"))
    kb.add_callback_button("+", color=VkKeyboardColor.POSITIVE,
                           payload=_payload(action="players_plus"))

    kb.add_line()
    kb.add_callback_button("Далее ▶", color=VkKeyboardColor.PRIMARY,
                           payload=_payload(action="players_next"))
    return kb


def build_slots_keyboard(page: int, selected: Set[Any], slots: List[Any]) -> VkKeyboard:
    """Клавиатура шага 2: выбор слотов с постраничной навигацией.

    slots — список объектов смен пользователя, полученный из Datasource.
    """
    kb = VkKeyboard(inline=True)

    if not slots:
        # Datasource не вернул смен — оставляем только возможность отправить заявку без них
        kb.add_callback_button("Отправить ✅", color=VkKeyboardColor.POSITIVE,
                               payload=_payload(action="submit"))
        return kb

    total_pages = max(1, (len(slots) - 1) // MAX_SLOTS + 1)
    page = page % total_pages
    start = page * MAX_SLOTS
    page_slots = slots[start:start + MAX_SLOTS]

    for i, shift in enumerate(page_slots):
        color = VkKeyboardColor.POSITIVE if shift_id(
            shift) in selected else VkKeyboardColor.SECONDARY
        kb.add_callback_button(shift_label(shift), color=color,
                               payload=_payload(action="slot_toggle", slot=shift_id(shift)))
        if i != len(page_slots) - 1:
            kb.add_line()  # каждый слот — на своей строке, так заметнее подсветка

    kb.add_line()
    kb.add_callback_button("◀", color=VkKeyboardColor.PRIMARY,
                           payload=_payload(action="slots_prev"))
    kb.add_callback_button(f"{page + 1}/{total_pages}", color=VkKeyboardColor.SECONDARY,
                           payload=_payload(action="noop"))
    kb.add_callback_button("▶", color=VkKeyboardColor.PRIMARY,
                           payload=_payload(action="slots_next"))
    kb.add_callback_button("Отправить ✅", color=VkKeyboardColor.POSITIVE,
                           payload=_payload(action="submit"))
    return kb


# ─────────────────────────────── БОТ ───────────────────────────────

class Bot:
    def __init__(self, token: str, group_id: int, datasource: Datasource) -> None:
        self.vk_session = vk_api.VkApi(token=token)
        self.vk = self.vk_session.get_api()
        self.long_poll = VkBotLongPoll(self.vk_session, group_id)
        self.datasource = datasource

    # ---- получение данных ----

    def _get_shifts(self, user_id: int) -> List[Any]:
        """
        Получает список смен через Datasource. Если у вашего get_shifts другая
        сигнатура (например без user_id, или с доп. фильтрами по дате) —
        поправьте только этот вызов.

        Любая ошибка (например недоступна БД) гасится здесь же, чтобы диалог
        не падал — пользователь увидит шаг с сообщением "смен не найдено".
        """
        try:
            return list(self.datasource.get_shifts(user_id))
        except Exception:
            log.exception(
                "Не удалось получить смены из Datasource для user_id=%s", user_id)
            return []

    # ---- низкоуровневая работа с сообщениями ----

    def send_step(self, peer_id: int, text: str, keyboard: VkKeyboard) -> int:
        """Отправляет новое сообщение с клавиатурой, возвращает его message_id."""
        return self.vk.messages.send(
            peer_id=peer_id,
            message=text,
            random_id=vk_api.utils.get_random_id(),
            keyboard=keyboard.get_keyboard(),
        )

    def edit_step(self, peer_id: int, message_id: int, text: str, keyboard: VkKeyboard) -> None:
        """Перерисовывает уже отправленное сообщение (тот же message_id)."""
        self.vk.messages.edit(
            peer_id=peer_id,
            message_id=message_id,
            message=text,
            keyboard=keyboard.get_keyboard(),
        )

    def answer_event(self, event_id: str, user_id: int, peer_id: int, snackbar: str = "") -> None:
        """Гасит "крутилку" у нажатой кнопки; опционально показывает всплывающую плашку."""
        kwargs = dict(event_id=event_id, user_id=user_id, peer_id=peer_id)
        if snackbar:
            # event_data передаём только когда он реально нужен — иначе VK
            # пытается распарсить "{}" как client_action и падает с ошибкой
            # "client_action has invalid type"
            kwargs["event_data"] = json.dumps(
                {"type": "show_snackbar", "text": snackbar})
        self.vk.messages.sendMessageEventAnswer(**kwargs)

    # ---- рендер шагов ----

    def render_players_step(self, state: UserState) -> None:
        text = f"Шаг 1 из 2. Количество игроков: {state.players}"
        self.edit_step(state.peer_id, state.message_id, text,
                       build_players_keyboard(state.players))

    def render_slots_step(self, state: UserState) -> None:
        if not state.slots:
            text = "Шаг 2 из 2. Свободных смен не найдено — можно отправить заявку без них."
        else:
            chosen_labels = [shift_label(s) for s in state.slots if shift_id(
                s) in state.slots_selected]
            chosen = ", ".join(chosen_labels) or "пока ничего не выбрано"
            text = f"Шаг 2 из 2. Выберите удобные слоты.\nВыбрано: {chosen}"
        self.edit_step(state.peer_id, state.message_id, text,
                       build_slots_keyboard(state.slot_page, state.slots_selected, state.slots))

    # ---- обработчики ----

    def handle_start(self, user_id: int, peer_id: int) -> None:
        state = reset_state(user_id)
        state.peer_id = peer_id
        # свежий список смен на каждый /start
        state.slots = self._get_shifts(user_id)
        text = f"Шаг 1 из 2. Количество игроков: {state.players}"
        state.message_id = self.send_step(
            peer_id, text, build_players_keyboard(state.players))

    def handle_callback(self, obj: dict) -> None:
        payload_data: dict = obj.get("payload") or {}
        action: Optional[str] = payload_data.get("action")
        user_id: int = obj.get("user_id")
        peer_id: int = obj.get("peer_id")
        event_id: str = obj.get("event_id")

        state = get_state(user_id)
        state.peer_id = peer_id
        if not state.message_id:
            # на случай рестарта бота: привязываемся к сообщению, на котором нажали кнопку
            state.message_id = obj.get("conversation_message_id")

        if action != "noop":
            if state.step == Step.PLAYERS:
                self._handle_players_action(state, action)
            elif state.step == Step.SLOTS:
                self._handle_slots_action(state, action, payload_data)

        self.answer_event(event_id, user_id, peer_id)

    def _handle_players_action(self, state: UserState, action: str) -> None:
        if action == "players_minus":
            state.players = max(MIN_PLAYERS, state.players - 1)
            self.render_players_step(state)
        elif action == "players_plus":
            state.players += 1
            self.render_players_step(state)
        elif action == "players_next":
            state.step = Step.SLOTS
            self.render_slots_step(state)

    def _handle_slots_action(self, state: UserState, action: str, payload_data: dict) -> None:
        total_pages = max(1, (len(state.slots) - 1) // MAX_SLOTS + 1)

        if action == "slot_toggle":
            sid = payload_data.get("slot")
            if sid in state.slots_selected:
                state.slots_selected.discard(sid)
            else:
                state.slots_selected.add(sid)
            self.render_slots_step(state)
        elif action == "slots_prev":
            state.slot_page = (state.slot_page - 1) % total_pages
            self.render_slots_step(state)
        elif action == "slots_next":
            state.slot_page = (state.slot_page + 1) % total_pages
            self.render_slots_step(state)
        elif action == "submit":
            self._handle_submit(state)

    def _handle_submit(self, state: UserState) -> None:
        chosen = [shift_label(s) for s in state.slots if shift_id(
            s) in state.slots_selected]
        slots_text = ", ".join(chosen) if chosen else "не выбраны"
        result_text = (
            "Заявка принята! ✅\n"
            f"Количество игроков: {state.players}\n"
            f"Выбранные слоты: {slots_text}"
        )
        self.vk.messages.edit(
            peer_id=state.peer_id,
            message_id=state.message_id,
            message=result_text,
            keyboard=VkKeyboard(inline=True).get_empty_keyboard(),
        )
        state.step = Step.DONE

    # ---- основной цикл ----

    def run(self) -> None:
        log.info("Бот запущен, жду события...")
        for event in self.long_poll.listen():
            try:
                if event.type == VkBotEventType.MESSAGE_NEW:
                    message = event.object.message
                    text = (message.get("text") or "").strip().lower()
                    if text == "/start":
                        self.handle_start(message.get(
                            "from_id"), message.get("peer_id"))

                elif event.type == VkBotEventType.MESSAGE_EVENT:
                    self.handle_callback(event.object)

            except Exception:
                log.exception("Ошибка при обработке события")


if __name__ == "__main__":
    # Подставьте сюда реальные зависимости вашего Datasource
    # (подключение к БД, конфиг и т.п.)
    datasource = Datasource(
        access_token=API_TOKEN,
        api_base=API_BASE
    )
    Bot(TOKEN, GROUP_ID, datasource).run()
