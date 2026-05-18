import http from './http'

export interface ScheduleDay {
  date: string
  is_holiday: boolean
  weekday: number
}

export interface ScheduleTask {
  id: number
  user_id: number
  person_name: string
  tech_direction: string
  color: string
  label: string
  start_date: string
  end_date: string
  work_days: number
}

export interface ScheduleBoard {
  start: string
  end: string
  today: string
  days: ScheduleDay[]
  tasks: ScheduleTask[]
  task_count: number
  project_count: number
}

export const getScheduleBoard = (start: string, end: string) =>
  http.get<ScheduleBoard>('/admin/work-schedules/board', { params: { start, end } })

export const createScheduleTask = (data: {
  user_id: number
  tech_direction: string
  color?: string
  label?: string
  start_date: string
  end_date: string
}) => http.post<{ id: number }>('/admin/work-schedules/tasks', data)

export const updateScheduleTask = (
  id: number,
  data: Partial<{
    user_id: number
    tech_direction: string
    color: string
    label: string
    start_date: string
    end_date: string
  }>
) => http.put(`/admin/work-schedules/tasks/${id}`, data)

export const deleteScheduleTask = (id: number) => http.delete(`/admin/work-schedules/tasks/${id}`)

export const toggleScheduleDay = (date: string) =>
  http.post<{ date: string; is_holiday: boolean }>('/admin/work-schedules/calendar/toggle', { date })
