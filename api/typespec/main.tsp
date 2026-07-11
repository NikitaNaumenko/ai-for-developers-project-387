import "@typespec/http";
import "@typespec/rest";

using TypeSpec.Http;

@service(#{ title: "Calendar API v2" })
@server("http://localhost:8080", "Локальный сервер разработки")
namespace Calendar.V2;

alias UuidString = string;

model EventType {
  @doc("Идентификатор типа события, заданный владельцем календаря.")
  @format("uuid")
  id: UuidString;

  @doc("Название типа события, которое показывается гостям.")
  @minLength(1)
  @maxLength(200)
  title: string;

  @doc("Описание типа события, которое показывается гостям перед бронированием.")
  description: string;

  @doc("Длительность события в минутах.")
  @minValue(1)
  durationMinutes: int32;
}

model CreateEventTypeRequest {
  @doc("Идентификатор типа события, заданный владельцем календаря.")
  @format("uuid")
  id: UuidString;

  @doc("Название типа события, которое показывается гостям.")
  @minLength(1)
  @maxLength(200)
  title: string;

  @doc("Описание типа события, которое показывается гостям перед бронированием.")
  description: string;

  @doc("Длительность события в минутах.")
  @minValue(1)
  durationMinutes: int32;
}

model AvailableSlot {
  @doc("Время начала слота в формате ISO 8601 date-time.")
  startsAt: utcDateTime;

  @doc("Время окончания слота в формате ISO 8601 date-time.")
  endsAt: utcDateTime;
}

model SlotsResponse {
  @doc("Тип события, для которого рассчитаны слоты.")
  @format("uuid")
  eventTypeId: UuidString;

  @doc("Начало рассчитанного сервером окна записи, включительно.")
  windowStartsAt: utcDateTime;

  @doc("Конец рассчитанного сервером окна записи, не включительно.")
  windowEndsAt: utcDateTime;

  @doc("Свободные слоты в окне записи. Сервер возвращает только слоты, доступные для бронирования.")
  slots: AvailableSlot[];
}

model Booking {
  @doc("Идентификатор бронирования, сгенерированный системой.")
  @format("uuid")
  id: UuidString;

  @doc("Идентификатор забронированного типа события.")
  @format("uuid")
  eventTypeId: UuidString;

  @doc("Название типа события для отображения в списке бронирований.")
  eventTypeTitle: string;

  @doc("Время начала бронирования в формате ISO 8601 date-time.")
  startsAt: utcDateTime;

  @doc("Время окончания бронирования в формате ISO 8601 date-time.")
  endsAt: utcDateTime;

  @doc("Имя гостя.")
  guestName: string;

  @doc("Email гостя.")
  @format("email")
  guestEmail: string;

  @doc("Необязательная заметка от гостя.")
  guestNote?: string;

  @doc("Время создания бронирования.")
  createdAt: utcDateTime;
}

model CreateBookingRequest {
  @doc("Тип события, выбранный гостем.")
  @format("uuid")
  eventTypeId: UuidString;

  @doc("Время начала выбранного слота. Значение должно совпадать с одним из свободных слотов, которые сервер вернул для текущего 14-дневного окна записи.")
  startsAt: utcDateTime;

  @doc("Имя гостя.")
  @minLength(1)
  @maxLength(200)
  guestName: string;

  @doc("Email гостя.")
  @format("email")
  guestEmail: string;

  @doc("Необязательная заметка от гостя.")
  guestNote?: string;
}

model ErrorResponse {
  @doc("Информация об ошибке.")
  error: ErrorBody;
}

model ErrorBody {
  @doc("Машиночитаемый код ошибки.")
  code: string;

  @doc("Человекочитаемое описание ошибки.")
  message: string;
}

@doc("Запрос успешно выполнен.")
model Ok<T> {
  @statusCode
  statusCode: 200;

  @body
  body: T;
}

@doc("Ресурс успешно создан.")
model Created<T> {
  @statusCode
  statusCode: 201;

  @body
  body: T;
}

@doc("Запрос содержит некорректные данные.")
model BadRequest {
  @statusCode
  statusCode: 400;

  @body
  body: ErrorResponse;
}

@doc("Запрошенный ресурс не найден.")
model NotFound {
  @statusCode
  statusCode: 404;

  @body
  body: ErrorResponse;
}

@doc("Запрос конфликтует с текущим состоянием системы.")
model Conflict {
  @statusCode
  statusCode: 409;

  @body
  body: ErrorResponse;
}

alias PublicError = BadRequest | NotFound;
alias BookingCreateError = BadRequest | NotFound | Conflict;

@route("/event-types")
@tag("Guest")
interface EventTypes {
  @doc("Возвращает список типов событий, доступных гостю для бронирования. Регистрация и авторизация не требуются.")
  @get
  listEventTypes(): Ok<EventType[]>;
}

@route("/event-types/{eventTypeId}/slots")
@tag("Guest")
interface Slots {
  @doc("Возвращает рассчитанные сервером свободные слоты для типа события. Слоты ограничены 14-дневным окном записи, начиная с текущей даты сервера.")
  @get
  listAvailableSlots(
    @doc("Идентификатор типа события.")
    @path
    @format("uuid")
    eventTypeId: UuidString,
  ): Ok<SlotsResponse> | PublicError;
}

@route("/bookings")
@tag("Guest")
interface Bookings {
  @doc("Создает бронирование гостя на свободный слот. Выбранное время начала должно совпадать со слотом, полученным из метода свободных слотов. Бронирование не может пересекаться с другим бронированием на то же время, даже если типы событий отличаются.")
  @post
  createBooking(@body body: CreateBookingRequest):
    | Created<Booking>
    | BookingCreateError;
}

@route("/admin/event-types")
@tag("Owner")
interface AdminEventTypes {
  @doc("Создает тип события для заранее заданного профиля владельца календаря. Авторизация в этом контракте не описывается.")
  @post
  createEventType(@body body: CreateEventTypeRequest):
    | Created<EventType>
    | BadRequest
    | Conflict;
}

@route("/admin/bookings/upcoming")
@tag("Owner")
interface AdminBookings {
  @doc("Возвращает предстоящие бронирования всех типов событий одним списком, отсортированным по возрастанию startsAt, для заранее заданного профиля владельца календаря.")
  @get
  listUpcomingBookings(): Ok<Booking[]>;
}
