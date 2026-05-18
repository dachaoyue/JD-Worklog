<template>
  <Shell full-width>
    <el-card class="schedule-card">
      <template #header>
        <div class="card-header">
          <b>开发排期计划表</b>
          <div class="header-actions">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="~"
              start-placeholder="开始"
              end-placeholder="结束"
              value-format="YYYY-MM-DD"
              :clearable="false"
              @change="loadBoard"
            />
            <el-button type="primary" @click="openNewScheduleDialog">新建排期</el-button>
            <el-button @click="loadBoard">刷新</el-button>
            <el-button @click="exportStaticSharePage">导出分享页</el-button>
          </div>
        </div>
      </template>

      <div
        v-loading="loading"
        class="gantt-split"
        @mouseup="onWrapMouseUp"
        @mouseleave="onWrapMouseUp"
      >
        <div class="gantt-fixed">
          <table class="gantt-table">
            <thead>
              <tr>
                <th class="col-person person-head-cell">
                  <el-dropdown trigger="click" @command="onPersonMenu">
                    <span class="person-head">
                      人员
                      <el-icon class="add-icon"><Plus /></el-icon>
                    </span>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item disabled>添加人员</el-dropdown-item>
                        <el-dropdown-item
                          v-for="u in usersNotOnBoard"
                          :key="'add-' + u.id"
                          :command="'add:' + u.id"
                        >
                          {{ u.nickname || u.username }}
                        </el-dropdown-item>
                        <el-dropdown-item v-if="usersNotOnBoard.length === 0" disabled>
                          暂无可添加用户
                        </el-dropdown-item>
                        <el-dropdown-item v-if="boardPersons.length" divided disabled>
                          移除人员
                        </el-dropdown-item>
                        <el-dropdown-item
                          v-for="p in boardPersons"
                          :key="'rm-' + p.user_id"
                          :command="'remove:' + p.user_id"
                        >
                          <span class="drop-remove-item">{{ p.person_name }}</span>
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </th>
                <th class="col-tech">技术方向</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="person in personRows"
                :key="'p-' + person.user_id"
                class="task-row"
              >
                <td class="col-person person-cell">
                  <span class="person-name">{{ person.person_name }}</span>
                </td>
                <td class="col-tech tech-cell">
                  <input
                    v-model="person.tech_direction"
                    class="tech-input"
                    placeholder="如：前端"
                    @blur="saveTechDirection(person)"
                  />
                </td>
              </tr>
              <tr v-if="personRows.length === 0">
                <td colspan="2" class="empty-row">点击「人员」添加</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="gantt-scroll">
          <table v-if="board" class="gantt-table">
            <thead>
              <tr>
                <th
                  v-for="day in board.days"
                  :key="day.date"
                  class="date-head"
                  :class="{ holiday: day.is_holiday }"
                  :title="day.is_holiday ? '节假日（点击改为工作日）' : '工作日（点击改为节假日）'"
                  @click="onToggleDay(day.date)"
                >
                  {{ formatHeadDate(day.date) }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="person in personRows"
                :key="'d-' + person.user_id"
                class="task-row"
              >
                <td
                  v-for="(day, dayIdx) in board.days"
                  :key="day.date"
                  class="date-cell"
                  :class="{
                    holiday: day.is_holiday && !cellHasSchedule(person, day.date),
                    preview: isPreviewCell(person, day.date),
                  }"
                  :style="cellBgStyle(person, day.date)"
                  @mousedown.prevent="onCellDown(person, day.date)"
                  @mouseenter="onCellEnter(person, day.date)"
                >
                  <template v-for="task in person.tasks" :key="task.id">
                    <div
                      v-if="isBarStart(task, day.date)"
                      class="bar-overlay"
                      :style="barSpanStyle(task, dayIdx)"
                    >
                      <input
                        v-model="task.label"
                        class="bar-input"
                        placeholder="项目说明"
                        :style="{ color: textColor(taskColor(task)) }"
                        @click.stop
                        @mousedown.stop
                        @blur="saveLabel(task)"
                      />
                      <button
                        type="button"
                        class="bar-del"
                        title="删除此段排期"
                        @mousedown.stop
                        @click.stop="onDeleteTask(task)">×</button>
                    </div>
                  </template>
                </td>
              </tr>
              <tr v-if="personRows.length === 0">
                <td :colspan="board.days.length" class="empty-row">
                  添加人员后，在其行的日期上拖拽选择范围，再选择颜色创建项目排期
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>


      <el-dialog
        v-model="newScheduleVisible"
        title="新建排期"
        width="480px"
        :close-on-click-modal="false"
        @closed="resetNewScheduleForm"
      >
        <el-form label-width="110px" class="new-schedule-form">
          <el-form-item label="人员" required>
            <el-select
              v-model="newScheduleForm.user_id"
              placeholder="请选择人员"
              filterable
              style="width: 100%"
            >
              <el-option
                v-for="u in users"
                :key="u.id"
                :label="u.nickname || u.username"
                :value="u.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="开始日期" required>
            <el-date-picker
              v-model="newScheduleForm.start_date"
              type="date"
              value-format="YYYY-MM-DD"
              placeholder="选择开始日期"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="工期天数" required>
            <el-input-number
              v-model="newScheduleForm.duration_days"
              :min="1"
              :max="366"
              controls-position="right"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="包括节假日">
            <el-radio-group v-model="newScheduleForm.include_holidays">
              <el-radio :value="true">是（连续日历天）</el-radio>
              <el-radio :value="false">否（仅计工作日）</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="排期颜色" required>
            <div class="color-field-wrap">
              <span
                class="color-selected-swatch"
                :style="{ background: newScheduleForm.color }"
              />
              <el-select
                v-model="newScheduleForm.color"
                class="color-select-inner"
                placeholder="选择颜色"
              >
                <el-option
                  v-for="c in COLOR_OPTIONS"
                  :key="c"
                  :value="c"
                  label=" "
                >
                  <span class="color-opt-swatch color-opt-swatch--option" :style="{ background: c }" />
                </el-option>
              </el-select>
            </div>
          </el-form-item>
          <el-form-item label="描述">
            <el-input
              v-model="newScheduleForm.label"
              type="textarea"
              :rows="2"
              placeholder="项目说明（可选）"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="newScheduleVisible = false">取消</el-button>
          <el-button type="primary" :loading="newScheduleSubmitting" @click="submitNewSchedule">
            确认
          </el-button>
        </template>
      </el-dialog>

      <el-dialog
        v-model="colorPickerVisible"
        title="选择项目颜色"
        width="360px"
        :close-on-click-modal="false"
        @closed="onColorDialogClosed"
      >
        <p v-if="pendingRange" class="pick-hint">
          为 <b>{{ pendingPersonName }}</b> 创建排期：
          {{ pendingRangeLabel }}
          <span v-if="pendingRange.segments.length > 1" class="pick-sub">
            （已跳过节假日，共 {{ pendingRange.segments.length }} 段）
          </span>
        </p>
        <div class="row-color-pick">
          <button
            v-for="c in COLOR_OPTIONS"
            :key="c"
            type="button"
            class="color-btn-lg"
            :style="{ background: c }"
            @click="confirmColor(c)"
          />
        </div>
      </el-dialog>
    </el-card>
  </Shell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import Shell from '../components/Shell.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAllUsers } from '../api/admin'
import {
  createScheduleTask,
  deleteScheduleTask,
  getScheduleBoard,
  toggleScheduleDay,
  updateScheduleTask,
  type ScheduleBoard,
  type ScheduleTask,
} from '../api/work_schedule'
import {
  downloadScheduleStaticPage,
  type StaticSchedulePayload,
} from '../utils/scheduleStaticExport'

const COLOR_OPTIONS = [
  '#409eff',
  '#67c23a',
  '#e6a23c',
  '#f56c6c',
  '#9b59b6',
  '#00d0b6',
  '#909399',
]

const CELL_W = 37
const DEFAULT_TECH = '前端'

type UserOpt = { id: number; username: string; nickname: string }
type BoardPerson = { user_id: number; person_name: string; tech_direction: string }
type PersonRow = BoardPerson & { tasks: ScheduleTask[] }

const loading = ref(false)
const board = ref<ScheduleBoard | null>(null)
const users = ref<UserOpt[]>([])
const dateRange = ref<[string, string]>(defaultRange())
const boardPersons = ref<BoardPerson[]>([])

type DateSegment = { start: string; end: string }
type PendingRange = { userId: number; start: string; end: string; segments: DateSegment[] }

const dragState = ref<{ userId: number; start: string; end: string } | null>(null)
const isDragging = ref(false)
const pendingRange = ref<PendingRange | null>(null)
const colorPickerVisible = ref(false)

type NewScheduleForm = {
  user_id: number | undefined
  start_date: string
  duration_days: number
  include_holidays: boolean
  color: string
  label: string
}

const newScheduleVisible = ref(false)
const newScheduleSubmitting = ref(false)
const newScheduleForm = ref<NewScheduleForm>({
  user_id: undefined,
  start_date: '',
  duration_days: 5,
  include_holidays: false,
  color: COLOR_OPTIONS[0],
  label: '',
})

const usersNotOnBoard = computed(() =>
  users.value.filter((u) => !boardPersons.value.some((p) => p.user_id === u.id))
)

const personRows = computed<PersonRow[]>(() =>
  boardPersons.value
    .map((p) => ({
      ...p,
      tasks: (board.value?.tasks ?? []).filter((t) => t.user_id === p.user_id),
    }))
    .sort((a, b) => a.user_id - b.user_id)
)

const pendingPersonName = computed(() => {
  if (!pendingRange.value) return ''
  const row = personRows.value.find((p) => p.user_id === pendingRange.value!.userId)
  return row?.person_name ?? ''
})

const pendingRangeLabel = computed(() => {
  const pr = pendingRange.value
  if (!pr) return ''
  if (pr.segments.length === 1) {
    return `${formatShort(pr.segments[0].start)} ~ ${formatShort(pr.segments[0].end)}`
  }
  return pr.segments.map((s) => `${formatShort(s.start)}~${formatShort(s.end)}`).join('、')
})

function parseYmd(s: string) {
  const [y, m, d] = s.split('-').map(Number)
  return new Date(y, m - 1, d)
}

function addCalendarDays(d: Date, n: number) {
  const r = new Date(d)
  r.setDate(r.getDate() + n)
  return r
}

function isHolidayDate(dateStr: string) {
  const hit = board.value?.days.find((d) => d.date === dateStr)
  if (hit) return hit.is_holiday
  const wd = parseYmd(dateStr).getDay()
  return wd === 0 || wd === 6
}

function calcScheduleSegments(
  start: string,
  durationDays: number,
  includeHolidays: boolean
): DateSegment[] {
  if (durationDays < 1) return []

  if (includeHolidays) {
    const end = fmt(addCalendarDays(parseYmd(start), durationDays - 1))
    return [{ start, end }]
  }

  const segments: DateSegment[] = []
  let cur = parseYmd(start)
  let counted = 0
  let segStart: string | null = null
  let segEnd: string | null = null
  let guard = 0

  while (counted < durationDays && guard < 500) {
    guard++
    const ds = fmt(cur)
    if (!isHolidayDate(ds)) {
      counted++
      if (!segStart) {
        segStart = ds
        segEnd = ds
      } else {
        segEnd = ds
      }
    } else if (segStart) {
      segments.push({ start: segStart, end: segEnd! })
      segStart = null
      segEnd = null
    }
    if (counted < durationDays) cur = addCalendarDays(cur, 1)
  }

  if (segStart) segments.push({ start: segStart, end: segEnd! })
  return segments
}

function ensurePersonOnBoard(userId: number) {
  if (boardPersons.value.some((p) => p.user_id === userId)) return
  const u = users.value.find((x) => x.id === userId)
  if (!u) return
  boardPersons.value.push({
    user_id: userId,
    person_name: u.nickname || u.username,
    tech_direction: DEFAULT_TECH,
  })
}

async function createTasksForUser(
  userId: number,
  segments: DateSegment[],
  color: string,
  label: string
) {
  ensurePersonOnBoard(userId)
  const person = boardPersons.value.find((p) => p.user_id === userId)
  const tech = (person?.tech_direction || '').trim() || DEFAULT_TECH
  for (const seg of segments) {
    await createScheduleTask({
      user_id: userId,
      tech_direction: tech,
      color,
      start_date: seg.start,
      end_date: seg.end,
      label: label.trim(),
    })
  }
}

function defaultRange(): [string, string] {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), 1)
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 0)
  return [fmt(start), fmt(end)]
}

function fmt(d: Date) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function formatShort(s: string) {
  const p = s.split('-')
  return p.length === 3 ? `${p[1]}-${p[2]}` : s
}

function formatHeadDate(s: string) {
  const p = s.split('-')
  return p.length === 3 ? `${Number(p[1])}/${Number(p[2])}` : s
}

function normalizeRange(start: string, end: string) {
  return start <= end ? { start, end } : { start: end, end: start }
}

function rangeContains(start: string, end: string, date: string) {
  const { start: a, end: b } = normalizeRange(start, end)
  return date >= a && date <= b
}

function rangeHasHoliday(start: string, end: string) {
  return (board.value?.days ?? []).some(
    (d) => d.date >= start && d.date <= end && d.is_holiday
  )
}

function splitRangeExcludingHolidays(start: string, end: string): DateSegment[] {
  const segments: DateSegment[] = []
  let segStart: string | null = null
  let segEnd: string | null = null

  for (const day of board.value?.days ?? []) {
    if (day.date < start || day.date > end) continue
    if (day.is_holiday) {
      if (segStart) {
        segments.push({ start: segStart, end: segEnd! })
        segStart = null
        segEnd = null
      }
      continue
    }
    if (!segStart) {
      segStart = day.date
      segEnd = day.date
    } else {
      segEnd = day.date
    }
  }
  if (segStart) segments.push({ start: segStart, end: segEnd! })
  return segments
}

function taskColor(task: ScheduleTask) {
  return task.color || COLOR_OPTIONS[0]
}

function taskAtDate(person: PersonRow, date: string) {
  return person.tasks.find((t) => date >= t.start_date && date <= t.end_date)
}

function isPreviewCell(person: PersonRow, date: string) {
  if (!isDragging.value || !dragState.value) return false
  if (dragState.value.userId !== person.user_id) return false
  return rangeContains(dragState.value.start, dragState.value.end, date)
}

function cellHasSchedule(person: PersonRow, date: string) {
  if (isPreviewCell(person, date)) return true
  return !!taskAtDate(person, date)
}

function cellBgStyle(person: PersonRow, date: string) {
  if (isPreviewCell(person, date)) {
    return { backgroundColor: 'rgba(64, 158, 255, 0.35)' }
  }
  const task = taskAtDate(person, date)
  if (task) return { backgroundColor: taskColor(task) }
  return {}
}

function isBarStart(task: ScheduleTask, date: string) {
  return date === task.start_date
}

function barSpanStyle(task: ScheduleTask, startIdx: number) {
  const days = board.value?.days ?? []
  let endIdx = days.findIndex((d) => d.date === task.end_date)
  if (endIdx < 0) endIdx = startIdx
  const span = Math.max(1, endIdx - startIdx + 1)
  return {
    width: `calc(${span} * ${CELL_W}px - 1px)`,
    background: taskColor(task),
  }
}

function textColor(bg: string) {
  const hex = bg.replace('#', '')
  if (hex.length !== 6) return '#fff'
  const r = parseInt(hex.slice(0, 2), 16)
  const g = parseInt(hex.slice(2, 4), 16)
  const b = parseInt(hex.slice(4, 6), 16)
  const lum = (0.299 * r + 0.587 * g + 0.114 * b) / 255
  return lum > 0.62 ? '#303133' : '#ffffff'
}

function mergePersonsFromTasks(tasks: ScheduleTask[]) {
  const seen = new Set(boardPersons.value.map((p) => p.user_id))
  for (const t of tasks) {
    if (seen.has(t.user_id)) continue
    seen.add(t.user_id)
    boardPersons.value.push({
      user_id: t.user_id,
      person_name: t.person_name,
      tech_direction: t.tech_direction || DEFAULT_TECH,
    })
  }
}


function exportStaticSharePage() {
  if (!board.value?.days?.length) {
    ElMessage.warning('当前无可导出数据，请先刷新排期')
    return
  }
  const [start, end] = dateRange.value
  const payload: StaticSchedulePayload = {
    title: '开发排期计划表',
    start,
    end,
    generatedAt: new Date().toLocaleString('zh-CN', { hour12: false }),
    days: board.value.days,
    persons: personRows.value.map((p) => ({
      user_id: p.user_id,
      person_name: p.person_name,
      tech_direction: p.tech_direction,
      tasks: p.tasks,
    })),
  }
  downloadScheduleStaticPage(payload)
  ElMessage.success('已生成静态分享页，请将下载的 HTML 文件发送给对方')
}

const loadBoard = async () => {
  if (!dateRange.value?.[0] || !dateRange.value?.[1]) return
  loading.value = true
  try {
    const { data } = await getScheduleBoard(dateRange.value[0], dateRange.value[1])
    board.value = data
    mergePersonsFromTasks(data.tasks)
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '加载失败')
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  const { data } = await getAllUsers()
  users.value = data
}

const onToggleDay = async (date: string) => {
  try {
    await toggleScheduleDay(date)
    await loadBoard()
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '切换失败')
  }
}

const onPersonMenu = (cmd: string) => {
  if (cmd.startsWith('add:')) addPerson(Number(cmd.slice(4)))
  else if (cmd.startsWith('remove:')) removePerson(Number(cmd.slice(7)))
}

const addPerson = (userId: number) => {
  const u = users.value.find((x) => x.id === userId)
  if (!u) return
  const name = u.nickname || u.username
  if (boardPersons.value.some((p) => p.user_id === userId)) {
    ElMessage.info('该人员已在列表中')
    return
  }
  boardPersons.value.push({ user_id: userId, person_name: name, tech_direction: DEFAULT_TECH })
  ElMessage.success('已添加，请在其行的日期格拖拽并选择颜色')
}

const removePerson = async (userId: number) => {
  const person = boardPersons.value.find((p) => p.user_id === userId)
  if (!person) return
  const tasks = board.value?.tasks.filter((t) => t.user_id === userId) ?? []
  try {
    if (tasks.length > 0) {
      await ElMessageBox.confirm(
        `移除「${person.person_name}」将同时删除其全部 ${tasks.length} 段排期，是否继续？`,
        '提示',
        { type: 'warning' }
      )
      for (const t of tasks) {
        await deleteScheduleTask(t.id)
      }
    }
    boardPersons.value = boardPersons.value.filter((p) => p.user_id !== userId)
    await loadBoard()
    ElMessage.success('已移除人员')
  } catch (err: any) {
    if (err === 'cancel' || err === 'close') return
    ElMessage.error(err.response?.data?.error || '移除失败')
  }
}

const saveTechDirection = async (person: PersonRow) => {
  const tech = (person.tech_direction || '').trim() || DEFAULT_TECH
  person.tech_direction = tech
  const bp = boardPersons.value.find((p) => p.user_id === person.user_id)
  if (bp) bp.tech_direction = tech
  for (const task of person.tasks) {
    if (task.tech_direction === tech) continue
    try {
      await updateScheduleTask(task.id, { tech_direction: tech })
      task.tech_direction = tech
    } catch (err: any) {
      ElMessage.error(err.response?.data?.error || '保存失败')
      await loadBoard()
      return
    }
  }
}

const onCellDown = (person: PersonRow, date: string) => {
  isDragging.value = true
  dragState.value = { userId: person.user_id, start: date, end: date }
}

const onCellEnter = (person: PersonRow, date: string) => {
  if (!isDragging.value || !dragState.value) return
  if (dragState.value.userId !== person.user_id) return
  dragState.value = { ...dragState.value, end: date }
}

const onWrapMouseUp = async () => {
  if (!isDragging.value || !dragState.value) {
    isDragging.value = false
    dragState.value = null
    return
  }
  const { userId, start, end } = dragState.value
  const { start: a, end: b } = normalizeRange(start, end)
  isDragging.value = false
  dragState.value = null

  let segments: DateSegment[] = [{ start: a, end: b }]
  if (rangeHasHoliday(a, b)) {
    try {
      await ElMessageBox.confirm(
        '所选日期范围内包含节假日，是否将节假日计入排期？',
        '提示',
        {
          confirmButtonText: '包括节假日',
          cancelButtonText: '不包括（跳过节假日）',
          distinguishCancelAndClose: true,
          type: 'info',
        }
      )
    } catch (action) {
      if (action === 'cancel') {
        segments = splitRangeExcludingHolidays(a, b)
        if (segments.length === 0) {
          ElMessage.warning('所选范围内没有工作日')
          return
        }
      } else {
        return
      }
    }
  }

  pendingRange.value = { userId, start: a, end: b, segments }
  colorPickerVisible.value = true
}


const openNewScheduleDialog = () => {
  newScheduleForm.value = {
    user_id: boardPersons.value[0]?.user_id ?? users.value[0]?.id,
    start_date: board.value?.today ?? fmt(new Date()),
    duration_days: 5,
    include_holidays: false,
    color: COLOR_OPTIONS[0],
    label: '',
  }
  newScheduleVisible.value = true
}

const resetNewScheduleForm = () => {
  newScheduleSubmitting.value = false
}

const submitNewSchedule = async () => {
  const form = newScheduleForm.value
  if (!form.user_id) {
    ElMessage.warning('请选择人员')
    return
  }
  if (!form.start_date) {
    ElMessage.warning('请选择开始日期')
    return
  }
  if (!form.duration_days || form.duration_days < 1) {
    ElMessage.warning('工期天数至少为 1')
    return
  }
  if (!form.color) {
    ElMessage.warning('请选择排期颜色')
    return
  }

  const segments = calcScheduleSegments(
    form.start_date,
    form.duration_days,
    form.include_holidays
  )
  if (segments.length === 0) {
    ElMessage.warning('未能计算出有效排期日期，请检查开始日期与工期')
    return
  }

  newScheduleSubmitting.value = true
  try {
    await createTasksForUser(form.user_id, segments, form.color, form.label)
    newScheduleVisible.value = false
    await loadBoard()
    if (segments.length > 1) {
      ElMessage.success(`已创建 ${segments.length} 段排期`)
    } else {
      ElMessage.success('排期已创建')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '创建失败')
  } finally {
    newScheduleSubmitting.value = false
  }
}

const confirmColor = async (color: string) => {
  if (!pendingRange.value) return
  const { userId, segments } = pendingRange.value
  try {
    await createTasksForUser(userId, segments, color, '')
    colorPickerVisible.value = false
    pendingRange.value = null
    await loadBoard()
    if (segments.length > 1) {
      ElMessage.success(`已创建 ${segments.length} 段排期，可在色块上输入说明`)
    } else {
      ElMessage.success('排期已创建，可在色块上输入说明')
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '创建失败')
  }
}

const onColorDialogClosed = () => {
  pendingRange.value = null
}

const saveLabel = async (task: ScheduleTask) => {
  try {
    await updateScheduleTask(task.id, { label: task.label ?? '' })
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '保存失败')
    await loadBoard()
  }
}

const onDeleteTask = async (task: ScheduleTask) => {
  const desc = task.label || '此段排期'
  try {
    await ElMessageBox.confirm(
      `删除「${desc}」排期？人员行保留。`,
      '提示',
      { type: 'warning' }
    )
    await deleteScheduleTask(task.id)
    ElMessage.success('已删除排期')
    await loadBoard()
  } catch (err: any) {
    if (err === 'cancel' || err === 'close') return
    ElMessage.error(err.response?.data?.error || '删除失败')
  }
}

onMounted(async () => {
  await loadUsers()
  await loadBoard()
})
</script>

<style scoped>
.schedule-card {
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.gantt-split {
  display: flex;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  max-width: 100%;
  user-select: none;
  overflow: hidden;
}

.gantt-fixed {
  flex-shrink: 0;
  background: #fff;
  z-index: 3;
  box-shadow: 4px 0 8px -4px rgba(0, 0, 0, 0.12);
}

.gantt-scroll {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  min-width: 0;
}

.gantt-table {
  border-collapse: collapse;
  font-size: 12px;
  table-layout: fixed;
}

.gantt-table th,
.gantt-table td {
  border: 1px solid #ebeef5;
  text-align: center;
  vertical-align: middle;
  white-space: nowrap;
  height: 40px;
  box-sizing: border-box;
}

.gantt-table thead th {
  background: #f5f7fa;
  font-weight: 600;
  padding: 6px 4px;
}

.col-person {
  width: 100px;
  min-width: 100px;
  padding: 6px 8px;
  text-align: left;
}

.col-tech {
  width: 80px;
  min-width: 80px;
  padding: 4px 6px;
}

.person-head-cell {
  cursor: pointer;
}

.person-head {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #409eff;
  cursor: pointer;
}

.person-head:hover {
  color: #66b1ff;
}

.add-icon {
  font-size: 14px;
}

.person-cell {
  text-align: left;
}

.tech-cell {
  text-align: left;
}

.person-name {
  font-weight: 500;
}

.tech-input {
  width: 100%;
  border: none;
  background: transparent;
  font-size: 12px;
  outline: none;
  padding: 2px 0;
}

.tech-input:focus {
  background: #fff;
  box-shadow: inset 0 0 0 1px #409eff;
  border-radius: 2px;
}

.drop-remove-item {
  color: #f56c6c;
}

.date-head,
.date-cell {
  width: 37px;
  min-width: 37px;
  padding: 0;
}

.date-head {
  cursor: pointer;
}

.date-head:hover {
  background: #ecf5ff;
}

.date-cell {
  position: relative;
  cursor: crosshair;
  overflow: visible;
}

.bar-overlay {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  z-index: 2;
  display: flex;
  align-items: center;
  box-sizing: border-box;
  pointer-events: auto;
}

.bar-input {
  flex: 1;
  min-width: 0;
  height: 100%;
  border: none;
  background: transparent;
  font-size: 11px;
  padding: 0 2px 0 4px;
  outline: none;
  text-align: left;
}

.bar-input::placeholder {
  opacity: 0.75;
}

.bar-del {
  flex-shrink: 0;
  width: 18px;
  height: 100%;
  border: none;
  background: rgba(0, 0, 0, 0.15);
  color: #fff;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0;
}

.bar-del:hover {
  background: rgba(0, 0, 0, 0.35);
}

.task-row:hover {
  background: #fafafa;
}

.holiday {
  background: #f4f4f5 !important;
}

.empty-row {
  padding: 16px;
  color: #909399;
  text-align: center;
}

.pick-hint {
  margin: 0 0 16px;
  font-size: 14px;
  color: #606266;
}

.pick-sub {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}

.row-color-pick {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
}

.color-btn-lg {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  border: 2px solid rgba(0, 0, 0, 0.1);
  cursor: pointer;
  padding: 0;
}

.color-btn-lg:hover {
  transform: scale(1.08);
}

.new-schedule-form {
  padding-right: 8px;
}

.color-field-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}

.color-select-inner {
  flex: 1;
  min-width: 0;
}

.color-select-inner :deep(.el-select__selected-item) {
  display: none;
}

.color-select-inner :deep(.el-select__placeholder) {
  color: #a8abb2;
}

.color-selected-swatch {
  flex-shrink: 0;
  display: block;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  border: 1px solid rgba(0, 0, 0, 0.15);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.2);
}

.color-opt-swatch {
  display: inline-block;
  border-radius: 4px;
  border: 1px solid rgba(0, 0, 0, 0.12);
  vertical-align: middle;
}

.color-opt-swatch--option {
  width: 28px;
  height: 28px;
}

.color-select-inner :deep(.el-select-dropdown__item) {
  display: flex;
  justify-content: center;
  padding: 8px 12px;
}
</style>
