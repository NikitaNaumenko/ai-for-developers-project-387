import {
  ActionIcon,
  Alert,
  AppShell,
  Badge,
  Button,
  Divider,
  Group,
  Loader,
  NavLink,
  NumberInput,
  Paper,
  ScrollArea,
  SimpleGrid,
  Stack,
  Text,
  TextInput,
  Textarea,
  Title,
  Tooltip,
} from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  IconCalendarBolt,
  IconCalendarCheck,
  IconCalendarEvent,
  IconCircleCheck,
  IconClock,
  IconDatabaseOff,
  IconListDetails,
  IconPlus,
  IconRefresh,
  IconUser,
  IconVideo,
} from '@tabler/icons-react';
import { FormEvent, useEffect, useMemo, useState } from 'react';
import {
  ApiError,
  AvailableSlot,
  Booking,
  CreateBookingRequest,
  CreateEventTypeRequest,
  EventType,
  calendarApi,
} from './api/client';

type Page = 'booking' | 'event-types' | 'bookings';

type BookingForm = {
  guestName: string;
  guestEmail: string;
  guestNote: string;
};

type EventTypeForm = {
  title: string;
  description: string;
  durationMinutes: number;
};

const initialBookingForm: BookingForm = {
  guestName: '',
  guestEmail: '',
  guestNote: '',
};

const initialEventTypeForm: EventTypeForm = {
  title: '',
  description: '',
  durationMinutes: 30,
};

export function App() {
  const [page, setPage] = useState<Page>('booking');
  const eventTypesQuery = useQuery({
    queryKey: ['event-types'],
    queryFn: calendarApi.listEventTypes,
  });

  return (
    <AppShell header={{ height: 64 }} navbar={{ width: 260, breakpoint: 'sm' }} padding={0}>
      <AppShell.Header>
        <div className="topbar">
          <Group gap="sm" wrap="nowrap">
            <div className="brand-mark">
              <IconCalendarBolt size={22} stroke={1.8} />
            </div>
            <div>
              <Title order={1} className="app-title">
                Calendar
              </Title>
              <Text size="xs" c="dimmed">
                Scheduling workspace
              </Text>
            </div>
          </Group>

          <Group gap="xs" wrap="nowrap">
            <StatusBadge isLoading={eventTypesQuery.isLoading} error={eventTypesQuery.error} />
            <Tooltip label="Refresh">
              <ActionIcon
                aria-label="Refresh"
                variant="subtle"
                color="gray"
                onClick={() => void eventTypesQuery.refetch()}
                loading={eventTypesQuery.isFetching}
              >
                <IconRefresh size={18} />
              </ActionIcon>
            </Tooltip>
          </Group>
        </div>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <Stack gap="lg" h="100%" justify="space-between">
          <Stack gap="xs">
            <Text size="xs" fw={700} c="dimmed" className="nav-kicker">
              Workspace
            </Text>
            <NavLink
              className="side-link"
              active={page === 'booking'}
              label="Booking page"
              leftSection={<IconCalendarEvent size={18} />}
              onClick={() => setPage('booking')}
            />
            <NavLink
              className="side-link"
              active={page === 'event-types'}
              label="Event types"
              leftSection={<IconListDetails size={18} />}
              onClick={() => setPage('event-types')}
            />
            <NavLink
              className="side-link"
              active={page === 'bookings'}
              label="Bookings"
              leftSection={<IconCalendarCheck size={18} />}
              onClick={() => setPage('bookings')}
            />
          </Stack>
          <div className="sidebar-profile">
            <div className="avatar-mark">CN</div>
            <div>
              <Text size="sm" fw={650}>
                Calendar
              </Text>
              <Text size="xs" c="dimmed">
                /calendar
              </Text>
            </div>
          </div>
        </Stack>
      </AppShell.Navbar>

      <AppShell.Main>
        <div className="mobile-nav">
          <Button
            size="xs"
            variant={page === 'booking' ? 'filled' : 'default'}
            leftSection={<IconCalendarEvent size={16} />}
            onClick={() => setPage('booking')}
          >
            Booking
          </Button>
          <Button
            size="xs"
            variant={page === 'event-types' ? 'filled' : 'default'}
            leftSection={<IconListDetails size={16} />}
            onClick={() => setPage('event-types')}
          >
            Event types
          </Button>
          <Button
            size="xs"
            variant={page === 'bookings' ? 'filled' : 'default'}
            leftSection={<IconCalendarCheck size={16} />}
            onClick={() => setPage('bookings')}
          >
            Bookings
          </Button>
        </div>
        <main className="workspace">
          {page === 'booking' && <BookingPage eventTypesQuery={eventTypesQuery} />}
          {page === 'event-types' && <EventTypesPage eventTypesQuery={eventTypesQuery} />}
          {page === 'bookings' && <BookingsPage />}
        </main>
      </AppShell.Main>
    </AppShell>
  );
}

function BookingPage({
  eventTypesQuery,
}: {
  eventTypesQuery: ReturnType<typeof useQuery<EventType[], Error>>;
}) {
  const queryClient = useQueryClient();
  const eventTypes = useMemo(() => normalizeEventTypes(eventTypesQuery.data), [eventTypesQuery.data]);
  const [selectedEventTypeId, setSelectedEventTypeId] = useState<string>('');
  const [selectedSlot, setSelectedSlot] = useState<AvailableSlot | null>(null);
  const [form, setForm] = useState<BookingForm>(initialBookingForm);

  useEffect(() => {
    if (!selectedEventTypeId && eventTypes[0]) {
      setSelectedEventTypeId(eventTypes[0].id);
    }
  }, [eventTypes, selectedEventTypeId]);

  useEffect(() => {
    setSelectedSlot(null);
  }, [selectedEventTypeId]);

  const selectedEventType = eventTypes.find((eventType) => eventType.id === selectedEventTypeId);
  const slotsQuery = useQuery({
    queryKey: ['slots', selectedEventTypeId],
    queryFn: () => calendarApi.listSlots(selectedEventTypeId),
    enabled: Boolean(selectedEventTypeId),
  });

  const slots = useMemo(() => normalizeSlots(slotsQuery.data?.slots), [slotsQuery.data?.slots]);
  const groupedSlots = useMemo(() => groupSlotsByDay(slots), [slots]);

  const bookingMutation = useMutation({
    mutationFn: calendarApi.createBooking,
    onSuccess: async () => {
      setForm(initialBookingForm);
      setSelectedSlot(null);
      await queryClient.invalidateQueries({ queryKey: ['bookings'] });
      notifications.show({
        color: 'dark',
        icon: <IconCircleCheck size={18} />,
        title: 'Booking confirmed',
        message: 'The guest booking has been created.',
      });
    },
    onError: (error: Error) => {
      notifications.show({
        color: 'red',
        title: 'Could not create booking',
        message: error.message,
      });
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!selectedEventType || !selectedSlot) {
      return;
    }

    const payload: CreateBookingRequest = {
      eventTypeId: selectedEventType.id,
      startsAt: selectedSlot.startsAt,
      guestName: form.guestName.trim(),
      guestEmail: form.guestEmail.trim(),
      guestNote: form.guestNote.trim() || undefined,
    };

    bookingMutation.mutate(payload);
  }

  return (
    <Stack gap="lg">
      <PageHeader
        eyebrow="Public link"
        title="Booking page"
        description="Select an event type, choose a time, and confirm the guest details."
      />

      {eventTypesQuery.error ? (
        <ErrorPanel error={eventTypesQuery.error} />
      ) : (
        <div className="booking-layout">
          <section className="event-column" aria-label="Event types">
            <div className="host-panel">
              <div className="avatar-mark avatar-mark-large">CN</div>
              <Text size="sm" c="dimmed">
                Calendar
              </Text>
              <Title order={2} className="host-title">
                Book a meeting
              </Title>
              <Text size="sm" c="dimmed">
                Choose a meeting type and pick an available time. Confirmation is sent instantly.
              </Text>
              <Group gap={8} mt="sm">
                <Badge variant="light" color="gray" leftSection={<IconVideo size={12} />}>
                  Online
                </Badge>
                <Badge variant="light" color="gray" leftSection={<IconClock size={12} />}>
                  Asia/Tashkent
                </Badge>
              </Group>
            </div>
            <Stack gap="sm">
              {eventTypesQuery.isLoading ? (
                <LoadingPanel label="Loading event types" />
              ) : eventTypes.length === 0 ? (
                <EmptyPanel label="No event types available" />
              ) : (
                eventTypes.map((eventType) => (
                  <button
                    className="event-type-card"
                    data-active={eventType.id === selectedEventTypeId || undefined}
                    key={eventType.id}
                    type="button"
                    onClick={() => setSelectedEventTypeId(eventType.id)}
                  >
                    <Group justify="space-between" align="flex-start" wrap="nowrap">
                      <div>
                        <Text fw={650}>{eventType.title}</Text>
                        <Text size="sm" c="dimmed" lineClamp={2}>
                          {eventType.description}
                        </Text>
                      </div>
                      <Badge variant="light" color="gray" leftSection={<IconClock size={12} />}>
                        {eventType.durationMinutes}m
                      </Badge>
                    </Group>
                  </button>
                ))
              )}
            </Stack>
          </section>

          <section className="scheduler-panel" aria-label="Available times">
            <div className="scheduler-heading">
              <div>
                <Text size="xs" fw={700} tt="uppercase" c="dimmed">
                  {selectedEventType ? `${selectedEventType.durationMinutes} min` : 'Select'}
                </Text>
                <Title order={2}>{selectedEventType?.title ?? 'Choose a time'}</Title>
              </div>
              <Badge variant="outline" color="gray">
                Asia/Tashkent
              </Badge>
            </div>

            <Divider />

            <ScrollArea className="slots-scroll" type="auto">
              {slotsQuery.isLoading ? (
                <LoadingPanel label="Loading slots" />
              ) : slotsQuery.error ? (
                <ErrorPanel error={slotsQuery.error} />
              ) : groupedSlots.length === 0 ? (
                <EmptyPanel label="No open slots" />
              ) : (
                <Stack gap="lg">
                  {groupedSlots.map((group) => (
                    <div key={group.day}>
                      <Text size="sm" fw={700} mb="xs">
                        {group.day}
                      </Text>
                      <div className="slot-grid">
                        {group.slots.map((slot) => (
                          <button
                            className="slot-button"
                            data-active={slot.startsAt === selectedSlot?.startsAt || undefined}
                            key={slot.startsAt}
                            type="button"
                            onClick={() => setSelectedSlot(slot)}
                          >
                            {formatTime(slot.startsAt)}
                          </button>
                        ))}
                      </div>
                    </div>
                  ))}
                </Stack>
              )}
            </ScrollArea>
          </section>

          <section className="details-panel" aria-label="Guest details">
            <form onSubmit={submit}>
              <Stack gap="md">
                <div>
                  <Title order={2} className="panel-title">
                    Guest details
                  </Title>
                  <Text size="sm" c="dimmed">
                    {selectedSlot ? formatFullDate(selectedSlot.startsAt) : 'No time selected'}
                  </Text>
                </div>
                <div className="selection-summary">
                  <Text size="xs" c="dimmed" fw={700} tt="uppercase">
                    Selection
                  </Text>
                  <Text size="sm" fw={650}>
                    {selectedEventType?.title ?? 'Event type'}
                  </Text>
                  <Text size="sm" c="dimmed">
                    {selectedSlot
                      ? `${formatDayLabel(selectedSlot.startsAt)} at ${formatTime(selectedSlot.startsAt)}`
                      : 'Pick a time to continue'}
                  </Text>
                </div>

                <TextInput
                  label="Name"
                  leftSection={<IconUser size={16} />}
                  value={form.guestName}
                  required
                  onChange={(event) => {
                    const { value } = event.currentTarget;
                    setForm((current) => ({ ...current, guestName: value }));
                  }}
                />
                <TextInput
                  label="Email"
                  type="email"
                  value={form.guestEmail}
                  required
                  onChange={(event) => {
                    const { value } = event.currentTarget;
                    setForm((current) => ({ ...current, guestEmail: value }));
                  }}
                />
                <Textarea
                  label="Note"
                  minRows={4}
                  value={form.guestNote}
                  onChange={(event) => {
                    const { value } = event.currentTarget;
                    setForm((current) => ({ ...current, guestNote: value }));
                  }}
                />
                <Button
                  type="submit"
                  leftSection={<IconCalendarCheck size={18} />}
                  loading={bookingMutation.isPending}
                  disabled={!selectedSlot || !form.guestName.trim() || !form.guestEmail.trim()}
                  fullWidth
                >
                  Confirm booking
                </Button>
              </Stack>
            </form>
          </section>
        </div>
      )}
    </Stack>
  );
}

function EventTypesPage({
  eventTypesQuery,
}: {
  eventTypesQuery: ReturnType<typeof useQuery<EventType[], Error>>;
}) {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<EventTypeForm>(initialEventTypeForm);
  const eventTypes = useMemo(() => normalizeEventTypes(eventTypesQuery.data), [eventTypesQuery.data]);

  const createMutation = useMutation({
    mutationFn: calendarApi.createEventType,
    onSuccess: async () => {
      setForm(initialEventTypeForm);
      await queryClient.invalidateQueries({ queryKey: ['event-types'] });
      notifications.show({
        color: 'dark',
        icon: <IconCircleCheck size={18} />,
        title: 'Event type created',
        message: 'The booking option is ready.',
      });
    },
    onError: (error: Error) => {
      notifications.show({
        color: 'red',
        title: 'Could not create event type',
        message: error.message,
      });
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const payload: CreateEventTypeRequest = {
      id: crypto.randomUUID(),
      title: form.title.trim(),
      description: form.description.trim(),
      durationMinutes: form.durationMinutes,
    };

    createMutation.mutate(payload);
  }

  return (
    <Stack gap="lg">
      <PageHeader
        eyebrow="Owner"
        title="Event types"
        description="Manage bookable meeting templates."
      />

      <div className="admin-layout">
        <section className="create-panel" aria-label="Create event type">
          <Paper withBorder radius="sm" p="md" className="surface-panel">
            <form onSubmit={submit}>
              <Stack gap="md">
                <Group justify="space-between" align="center">
                  <Title order={2} className="panel-title">
                    New event type
                  </Title>
                  <Badge variant="light" color="gray">
                    POST /admin/event-types
                  </Badge>
                </Group>
                <TextInput
                  label="Title"
                  value={form.title}
                  required
                  maxLength={200}
                  onChange={(event) => {
                    const { value } = event.currentTarget;
                    setForm((current) => ({ ...current, title: value }));
                  }}
                />
                <Textarea
                  label="Description"
                  minRows={4}
                  value={form.description}
                  required
                  onChange={(event) => {
                    const { value } = event.currentTarget;
                    setForm((current) => ({
                      ...current,
                      description: value,
                    }));
                  }}
                />
                <NumberInput
                  label="Duration"
                  min={1}
                  suffix=" min"
                  value={form.durationMinutes}
                  onChange={(value) =>
                    setForm((current) => ({
                      ...current,
                      durationMinutes: typeof value === 'number' ? value : 30,
                    }))
                  }
                />
                <Button
                  type="submit"
                  leftSection={<IconPlus size={18} />}
                  loading={createMutation.isPending}
                  disabled={!form.title.trim() || !form.description.trim()}
                  fullWidth
                >
                  Create event type
                </Button>
              </Stack>
            </form>
          </Paper>
        </section>

        <section aria-label="Event type list">
          {eventTypesQuery.error ? (
            <ErrorPanel error={eventTypesQuery.error} />
          ) : eventTypesQuery.isLoading ? (
            <LoadingPanel label="Loading event types" />
          ) : eventTypes.length === 0 ? (
            <EmptyPanel label="No event types yet" />
          ) : (
            <SimpleGrid cols={{ base: 1, md: 2 }} spacing="md">
              {eventTypes.map((eventType) => (
                <Paper withBorder radius="sm" p="md" key={eventType.id} className="surface-panel">
                  <Stack gap="sm">
                    <Group justify="space-between" align="flex-start" wrap="nowrap">
                      <div>
                        <Text fw={700}>{eventType.title}</Text>
                        <Text size="sm" c="dimmed" lineClamp={3}>
                          {eventType.description}
                        </Text>
                      </div>
                      <Badge variant="light" color="gray">
                        {eventType.durationMinutes}m
                      </Badge>
                    </Group>
                    <Text size="xs" c="dimmed" className="mono">
                      {eventType.id}
                    </Text>
                  </Stack>
                </Paper>
              ))}
            </SimpleGrid>
          )}
        </section>
      </div>
    </Stack>
  );
}

function BookingsPage() {
  const bookingsQuery = useQuery({
    queryKey: ['bookings'],
    queryFn: calendarApi.listUpcomingBookings,
  });
  const bookings = useMemo(() => normalizeBookings(bookingsQuery.data), [bookingsQuery.data]);

  return (
    <Stack gap="lg">
      <PageHeader
        eyebrow="Owner"
        title="Upcoming bookings"
        description="Confirmed guest meetings across all event types."
      />

      {bookingsQuery.error ? (
        <ErrorPanel error={bookingsQuery.error} />
      ) : bookingsQuery.isLoading ? (
        <LoadingPanel label="Loading bookings" />
      ) : bookings.length === 0 ? (
        <EmptyPanel label="No upcoming bookings" />
      ) : (
        <Stack gap="sm">
          {bookings.map((booking) => (
            <Paper withBorder radius="sm" p="md" key={booking.id} className="surface-panel">
              <Group justify="space-between" align="flex-start" gap="lg">
                <Group gap="md" align="flex-start">
                  <div className="date-tile">
                    <Text size="xs" fw={700} tt="uppercase">
                      {formatMonth(booking.startsAt)}
                    </Text>
                    <Text fw={800} size="xl">
                      {formatDay(booking.startsAt)}
                    </Text>
                  </div>
                  <div>
                    <Text fw={700}>{booking.eventTypeTitle}</Text>
                    <Text size="sm" c="dimmed">
                      {formatFullDate(booking.startsAt)} - {formatTime(booking.endsAt)}
                    </Text>
                    <Text size="sm" mt={6}>
                      {booking.guestName} · {booking.guestEmail}
                    </Text>
                    {booking.guestNote ? (
                      <Text size="sm" c="dimmed" mt={4} lineClamp={2}>
                        {booking.guestNote}
                      </Text>
                    ) : null}
                  </div>
                </Group>
                <Badge variant="light" color="gray">
                  Confirmed
                </Badge>
              </Group>
            </Paper>
          ))}
        </Stack>
      )}
    </Stack>
  );
}

function PageHeader({
  eyebrow,
  title,
  description,
}: {
  eyebrow: string;
  title: string;
  description: string;
}) {
  return (
    <div className="page-header">
      <Text size="xs" fw={700} tt="uppercase" c="dimmed">
        {eyebrow}
      </Text>
      <Title order={1}>{title}</Title>
      <Text c="dimmed">{description}</Text>
    </div>
  );
}

function StatusBadge({ isLoading, error }: { isLoading: boolean; error: Error | null }) {
  if (isLoading) {
    return (
      <Badge variant="light" color="gray">
        Connecting
      </Badge>
    );
  }

  if (error) {
    return (
      <Badge variant="light" color="red" leftSection={<IconDatabaseOff size={12} />}>
        Offline
      </Badge>
    );
  }

  return (
    <Badge variant="light" color="gray" leftSection={<IconCircleCheck size={12} />}>
      Online
    </Badge>
  );
}

function LoadingPanel({ label }: { label: string }) {
  return (
    <div className="state-panel">
      <Loader size="sm" />
      <Text size="sm" c="dimmed">
        {label}
      </Text>
    </div>
  );
}

function EmptyPanel({ label }: { label: string }) {
  return (
    <div className="state-panel">
      <Text size="sm" c="dimmed">
        {label}
      </Text>
    </div>
  );
}

function ErrorPanel({ error }: { error: Error }) {
  const message =
    error instanceof ApiError ? `${error.code}: ${error.message}` : 'The API is not reachable.';

  return (
    <Alert color="red" variant="light" icon={<IconDatabaseOff size={18} />}>
      {message}
    </Alert>
  );
}

function normalizeEventTypes(eventTypes: EventType[] | undefined): EventType[] {
  return [...(eventTypes ?? [])].sort((left, right) => left.title.localeCompare(right.title));
}

function normalizeSlots(slots: AvailableSlot[] | undefined): AvailableSlot[] {
  return [...(slots ?? [])].sort(
    (left, right) => new Date(left.startsAt).getTime() - new Date(right.startsAt).getTime(),
  );
}

function normalizeBookings(bookings: Booking[] | undefined): Booking[] {
  return [...(bookings ?? [])].sort(
    (left, right) => new Date(left.startsAt).getTime() - new Date(right.startsAt).getTime(),
  );
}

function groupSlotsByDay(slots: AvailableSlot[]) {
  const groups = new Map<string, AvailableSlot[]>();

  slots.forEach((slot) => {
    const day = formatDayLabel(slot.startsAt);
    groups.set(day, [...(groups.get(day) ?? []), slot]);
  });

  return Array.from(groups, ([day, daySlots]) => ({ day, slots: daySlots }));
}

function formatDayLabel(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    weekday: 'long',
    month: 'short',
    day: 'numeric',
  }).format(new Date(value));
}

function formatFullDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}

function formatTime(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}

function formatMonth(value: string): string {
  return new Intl.DateTimeFormat(undefined, { month: 'short' }).format(new Date(value));
}

function formatDay(value: string): string {
  return new Intl.DateTimeFormat(undefined, { day: '2-digit' }).format(new Date(value));
}
