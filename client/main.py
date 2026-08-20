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

import csv
import io
from dotenv import load_dotenv
import json
import logging
from dataclasses import dataclass, field
from enum import Enum
import os
from typing import Any, Callable, Dict, List, Optional, Protocol, Set

import arrow
import requests
import vk_api
from vk_api.bot_longpoll import VkBotEventType, VkBotLongPoll
from vk_api.keyboard import VkKeyboard, VkKeyboardColor

# ваш класс — поправьте путь импорта при необходимости
from datasource import Datasource, Demand

# ─────────────────────────────── НАСТРОЙКИ ───────────────────────────────


load_dotenv()  # reads variables from a .env file and sets them in os.environ
# токен сообщества (Bot API, права: messages)
TOKEN = os.environ["VK_TOKEN"]
GROUP_ID = int(os.environ["VK_GROUP_ID"])
API_BASE = os.environ["API_BASE"].rstrip("/")  # e.g. http://localhost:8000
API_TOKEN = os.environ["CLIENT_API_KEY"]           # числовой id сообщества

MAX_SLOTS = 5      # сколько кнопок-слотов показывается на одной странице
MIN_PLAYERS = 1     # минимальное количество игроков
# команды, которые запускают диалог
START_COMMANDS = ["/start", "начать", "старт"]
ADD_DEMAND_COMMAND = "Записаться"
SHOW_DEMAND_COMMAND = "Мои заявки"
GET_DEMANDS_COMMAND = "Выгрузить заявки"
DELETE_DEMAND_COMMAND = "Удалить заявку"
ADMIN_IDS: set[int] = {947119, 100456345}

logging.basicConfig(level=logging.INFO,
                    format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)


def slot_id(shift: Any) -> Any:
    """
    Достаёт идентификатор смены — он используется как ключ в payload кнопки
    и в slots_selected. Если у объекта, который возвращает ваш Datasource,
    поле называется иначе (например shift_id вместо id) — поправьте только
    эту строку.
    """
    return shift["id"]


def slot_label(shift: Any) -> str:
    """
    Достаёт человекочитаемое название смены для подписи на кнопке
    (например "Пн 18:00–20:00"). Поправьте под реальное поле вашего Shift —
    например shift.name или shift["title"], если это dict.
    """
    return shift["name"]


def get_next_week(short: bool = True) -> str:
    next_week = arrow.now().shift(days=1-arrow.now().isoweekday()+7)
    if short:
        return next_week.format("YYYY-MM-DD")

    return next_week.isoformat()


# ─────────────────────────────── СОСТОЯНИЕ ───────────────────────────────


class Step(str, Enum):
    PLAYERS = "players"
    SLOTS = "slots"
    COMMENT = "comment"
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
    first_name: str = "Игрок"
    last_name: str = ""
    comment: str = ""
    username: str = "NONE"  # можно достать из профиля VK, если нужно


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


def build_keyboard(user_id: int) -> str:
  # из env/конфига
    keyboard = VkKeyboard(one_time=False)

    keyboard.add_button(ADD_DEMAND_COMMAND, color=VkKeyboardColor.PRIMARY)
    keyboard.add_button(SHOW_DEMAND_COMMAND, color=VkKeyboardColor.SECONDARY)
    keyboard.add_line()
    keyboard.add_button(DELETE_DEMAND_COMMAND,
                        color=VkKeyboardColor.NEGATIVE)

    if user_id in ADMIN_IDS:
        keyboard.add_line()
        keyboard.add_button(GET_DEMANDS_COMMAND,
                            color=VkKeyboardColor.NEGATIVE)

    return keyboard.get_keyboard()


def build_players_keyboard(players: int) -> VkKeyboard:
    """Клавиатура шага 1: выбор количества игроков."""
    kb = VkKeyboard(inline=True)

    kb.add_callback_button("−", color=VkKeyboardColor.NEGATIVE,
                           payload=_payload(action="players_minus"))
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
        prefix = "✅ " if slot_id(shift) in selected else ""
        kb.add_callback_button("{0}{1}".format(prefix, slot_label(shift)), color=VkKeyboardColor.SECONDARY,
                               payload=_payload(action="slot_toggle", slot=slot_id(shift)))

        # color = VkKeyboardColor.POSITIVE if slot_id(
        #     shift) in selected else VkKeyboardColor.SECONDARY
        # kb.add_callback_button(slot_label(shift), color=color,
        #                        payload=_payload(action="slot_toggle", slot=slot_id(shift)))
        if i != len(page_slots) - 1:
            kb.add_line()  # каждый слот — на своей строке, так заметнее подсветка

    kb.add_line()
    kb.add_callback_button("◀", color=VkKeyboardColor.PRIMARY,
                           payload=_payload(action="slots_prev"))
    kb.add_callback_button(f"{page + 1}/{total_pages}", color=VkKeyboardColor.SECONDARY,
                           payload=_payload(action="noop"))
    kb.add_callback_button("▶", color=VkKeyboardColor.PRIMARY,
                           payload=_payload(action="slots_next"))
    kb.add_callback_button("Далее", color=VkKeyboardColor.POSITIVE,
                           payload=_payload(action="submit"))
    return kb


# ─────────────────────────────── БОТ ───────────────────────────────

@dataclass(frozen=True, slots=True)
class MessageContext:
    from_id: int
    peer_id: int
    text: str


class CommandHandler(Protocol):
    def __call__(self, ctx: MessageContext) -> None: ...


class Bot:
    def __init__(self, token: str, group_id: int, datasource: Datasource, allowed_days: list[int]) -> None:
        self.http = requests.Session()
        self.http.headers.pop('user-agent')

        self.allowed_days = allowed_days
        self.vk_session = vk_api.VkApi(token=token)
        self.upload = vk_api.VkUpload(self.vk_session)
        self.vk = self.vk_session.get_api()
        self.long_poll = VkBotLongPoll(self.vk_session, group_id)
        self.datasource = datasource

        self._command_handlers: dict[str, Callable[[MessageContext], None]] = {
            **{cmd.lower(): self.handle_start for cmd in START_COMMANDS},
            SHOW_DEMAND_COMMAND.lower(): self.handle_show_my_demands,
            ADD_DEMAND_COMMAND.lower(): self.handle_add_demand,
            GET_DEMANDS_COMMAND.lower(): self.handle_get_all_demands,
            DELETE_DEMAND_COMMAND.lower(): self.handle_delete_demand,
        }

    def _dispatch(self, event) -> None:
        match event.type:
            case VkBotEventType.MESSAGE_NEW:
                self._handle_message_new(event.object.message)
            case VkBotEventType.MESSAGE_EVENT:
                self.handle_callback(event.object)

    def _handle_message_new(self, message: dict) -> None:
        ctx = MessageContext(
            from_id=message.get("from_id"),
            peer_id=message.get("peer_id"),
            text=(message.get("text") or "").strip().lower(),
        )
        handler = self._command_handlers.get(ctx.text)
        if handler:
            handler(ctx)
        elif get_state(ctx.from_id).step == Step.COMMENT:
            self._handle_comment(ctx)

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
            return list(self.datasource.get_slots(user_id, for_week=get_next_week(short=True)))
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
        log.info(
            f"Rendering selected slots for user_id={state.peer_id}: {state.slots_selected}")
        if not state.slots:
            text = "Шаг 2 из 2. Свободных смен не найдено — можно отправить заявку без них."
        else:
            chosen_labels = [slot_label(s) for s in state.slots if slot_id(
                s) in state.slots_selected]
            chosen = ", ".join(chosen_labels) or "пока ничего не выбрано"
            text = f"Шаг 2 из 2. Выберите удобные слоты.\nВыбрано: {chosen}"
        self.edit_step(state.peer_id, state.message_id, text,
                       build_slots_keyboard(state.slot_page, state.slots_selected, state.slots))

    # ---- обработчики ----

    def handle_add_demand(self, context: MessageContext) -> None:
        if arrow.now().isoweekday() not in self.allowed_days:
            text = "Прием заявок на следующую неделю окончен! Проверьте свои заявки через кнопку «Мои заявки»."
            self.vk.messages.send(
                user_id=context.from_id,
                message=text,
                random_id=vk_api.utils.get_random_id(),
            )
            return

        user_id = context.from_id
        state = reset_state(user_id)
        log.info(
            f"Start add demand conversation for user_id={user_id}, selected slots={state.slots_selected}")

        state.peer_id = context.peer_id
        # свежий список смен на каждый /start
        state.slots = self._get_shifts(user_id)
        text = f"Шаг 1 из 2. Количество игроков: {state.players}"
        state.message_id = self.send_step(
            context.peer_id, text, build_players_keyboard(state.players))

    def handle_show_my_demands(self, context: MessageContext) -> None:
        user_id = context.from_id
        # получаем актуальные заявки через Datasource
        demands = self.datasource.get_demands(
            user_id, for_week=get_next_week())
        log.info(
            f"Receive {len(demands)} demands for user {user_id}")

        text = "У вас пока нет заявок на следующую неделю."
        if demands:
            demand = demands[0]
            slots = demand.slots
            text = "Ваши заявки на следующую неделю: \n{} человек\nвременные слоты: {}\nкомментарий: {}".format(
                demand.players_count, ", ".join([slot_label(s) for s in slots]), demand.comment or "не указан")

        self.vk.messages.send(
            user_id=context.from_id,
            message=text,
            random_id=vk_api.utils.get_random_id(),
        )

    def handle_get_all_demands(self, context: MessageContext) -> None:
        if context.from_id not in ADMIN_IDS:
            return

        demands = self.datasource.get_all_demands(for_week=get_next_week())

        text = "На следующую неделю нет заявок"
        if demands:
            text = f"Выгружено {len(demands)} заявок."

            slots = self.datasource.get_slots(
                context.from_id, for_week=get_next_week(short=True))
            demands_table = self._create_flat_csv(demands=demands, slots=slots)
            self.send_generated_document(
                peer_id=context.from_id, filename="demands.csv", rows=demands_table)

        self.vk.messages.send(
            user_id=context.from_id,
            message=text,
            random_id=vk_api.utils.get_random_id(),
        )

    def handle_delete_demand(self, context: MessageContext) -> None:
        self.datasource.delete_demands(
            context.from_id, for_week=get_next_week(), user_id=context.from_id)

        self.vk.messages.send(
            user_id=context.from_id,
            message="Ваши заявки на следующую неделю удалены.",
            random_id=vk_api.utils.get_random_id(),
        )

    def send_generated_document(self, peer_id: int, filename: str, rows: list[list[str]]) -> None:
        buffer = io.StringIO()
        writer = csv.writer(buffer)
        writer.writerows(rows)

        file_bytes = buffer.getvalue().encode("utf-8-sig")

        upload_url = self.vk.docs.getMessagesUploadServer(
            peer_id=peer_id, type="doc"
        )["upload_url"]

        session = requests.Session()
        # VK's upload servers can reject the default requests UA
        session.headers.pop("User-Agent", None)

        print("FILENAME:", repr(filename))
        print("FILE BYTES LENGTH:", len(file_bytes))
        print("FILE BYTES PREFIX:", repr(file_bytes[:100]))

        for attempt in range(3):
            try:
                raw = session.post(
                    upload_url,
                    files={
                        "file": (
                            filename,
                            file_bytes,
                            "text/csv",
                        )
                    },
                    timeout=(10, 30),
                )

                print("UPLOAD URL:", upload_url)
                print("STATUS:", raw.status_code)
                print("CONTENT-TYPE:", raw.headers.get("Content-Type"))
                print("CONTENT-LENGTH:", raw.headers.get("Content-Length"))
                print("RESPONSE HEADERS:", dict(raw.headers))
                print("RESPONSE BODY:", repr(raw.text[:2000]))

                raw.raise_for_status()
                break
            except requests.exceptions.RequestException as e:
                log.warning(
                    "VK upload server request failed (attempt %d/3): %s", attempt + 1, e)
                if attempt == 2:
                    raise RuntimeError(
                        f"VK upload server request failed after 3 attempts: {e}"
                    ) from e
                continue
        if not raw.text.strip():
            raise RuntimeError(
                f"VK upload server returned an empty response "
                f"(status={raw.status_code})"
            )

        try:
            upload_response = raw.json()
        except requests.exceptions.JSONDecodeError as e:
            raise RuntimeError(
                "VK upload server returned a non-JSON response: "
                f"status={raw.status_code}, "
                f"content_type={raw.headers.get('Content-Type')!r}, "
                f"body={raw.text[:2000]!r}"
            ) from e

        if "file" not in upload_response:
            raise RuntimeError(
                f"VK upload server returned no file token: {upload_response}")

        doc = self.vk.docs.save(file=upload_response["file"], title=filename)
        doc_data = doc["doc"]
        attachment = f"doc{doc_data['owner_id']}_{doc_data['id']}"

        self.vk.messages.send(
            peer_id=peer_id,
            attachment=attachment,
            random_id=vk_api.utils.get_random_id(),
        )

    def _create_flat_csv(self, demands: list[Demand], slots: list[dict[str, Any]]) -> list[list[str]]:
        slots_names = [slot_label(slot) for slot in slots]

        header = ["link", "name", "username",
                  "players_count", "comment", *slots_names, ]
        rows: list[list[str]] = [header]
        for demand in demands:
            row: list[str] = [
                "https://vk.ru/id{}".format(demand.vk_id),
                "{0} {1}".format(demand.first_name, demand.last_name),
                demand.vk_username if demand.vk_username != "NONE" else "NONE",
                demand.players_count,
                demand.comment
            ]

            current_slot_ids = set([slot_id(slot) for slot in demand.slots])
            for slot in slots:
                slot_txt = ""
                if slot_id(slot) in current_slot_ids:
                    slot_txt = "1"
                row.append(slot_txt)
            rows.append(row)
        return rows

    def handle_start(self, context: MessageContext) -> None:
        self.vk.messages.send(
            user_id=context.from_id,
            message="Выберите действие:",
            keyboard=build_keyboard(context.from_id),
            random_id=vk_api.utils.get_random_id(),
        )

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

        # Пользователь нажал кнопку из уже завершённой сессии (например, скроллит
        # историю и кликает старое сообщение) — просто гасим крутилку и выходим
        if state.step == Step.DONE:
            self.answer_event(event_id, user_id, peer_id)
            return

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
        """Нажатие 'Далее' на шаге слотов → переходим к шагу комментария."""
        state.step = Step.COMMENT
        # Убираем inline-клавиатуру и просим написать комментарий обычным сообщением.
        # Пользователь отвечает текстом — это message_new, не callback.
        self.vk.messages.edit(
            peer_id=state.peer_id,
            message_id=state.message_id,
            message="Шаг 3 из 3. Напишите комментарий к заявке (или отправьте «-», чтобы пропустить).",
            keyboard=VkKeyboard(inline=True).get_empty_keyboard(),
        )

    def _handle_comment(self, context: MessageContext) -> None:
        """Получаем текстовый комментарий и отправляем итоговую заявку."""
        state = get_state(context.from_id)
        text = context.text.replace("\n", " || ").strip()
        state.comment = "" if text.strip() == "-" else text.strip()

        chosen = [s for s in state.slots
                  if slot_id(s) in state.slots_selected]
        slots_text = ", ".join([slot_label(s)
                                for s in chosen]) if chosen else "не выбраны"
        comment_text = state.comment if state.comment else "не указан"

        result_text = (
            "Заявка принята! ✅\n"
            f"Количество игроков: {state.players}\n"
            f"Выбранные слоты: {slots_text}\n"
            f"Комментарий: {comment_text}"
        )
        self.vk.messages.send(
            peer_id=state.peer_id,
            message=result_text,
            random_id=vk_api.utils.get_random_id(),
            keyboard=build_keyboard(state.peer_id),
        )

        # Отправляем заявку через Datasource
        user = self.vk.users.get(
            user_id=state.peer_id, fields="first_name,last_name,screen_name")
        if not user:
            log.warning(
                f"Не удалось получить данные пользователя VK для user_id={state.peer_id}")
            first_name = "Игрок"
            last_name = ""
            username = "NONE"
        else:
            first_name = user[0].get("first_name", "Игрок")
            last_name = user[0].get("last_name", "")
            username = user[0].get("screen_name", "NONE")

        log.info(f"Sending post request {first_name} {last_name}, {username}")
        demand = Demand(
            vk_id=state.peer_id,
            slots=chosen,
            players_count=state.players,
            first_name=first_name,
            last_name=last_name,
            vk_username=username,
            comment=state.comment,
            for_week=get_next_week(short=False)
        )
        self.datasource.post_demands(demand)

        # Очищаем состояние
        state.slots_selected = set()
        state.slots = []
        state.comment = ""
        state.step = Step.DONE

    # ---- основной цикл ----

    def run(self) -> None:
        log.info("Бот запущен, жду события...")
        for event in self.long_poll.listen():
            try:
                self._dispatch(event)
            except Exception:
                log.exception("Ошибка при обработке события: %s", event.type)


if __name__ == "__main__":
    # Подставьте сюда реальные зависимости вашего Datasource
    # (подключение к БД, конфиг и т.п.)
    datasource = Datasource(
        access_token=API_TOKEN,
        api_base=API_BASE
    )
    Bot(token=TOKEN,
        group_id=GROUP_ID,
        datasource=datasource,
        allowed_days=[1, 2, 3, 4, 5, 6, 7]
        ).run()
