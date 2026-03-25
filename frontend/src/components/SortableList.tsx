'use client'

import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical } from 'lucide-react'

export interface SortableItemRenderProps {
  handleProps: React.HTMLAttributes<HTMLDivElement>
  setNodeRef: (element: HTMLElement | null) => void
  style: React.CSSProperties
}

export interface SortableDndProviderProps<T> {
  items: T[]
  itemId: (item: T) => string
  onReorder: (ids: string[]) => void | Promise<void>
  disabled?: boolean
  children: React.ReactNode
}

export interface SortableListProps<T> {
  items: T[]
  itemId: (item: T) => string
  onReorder: (ids: string[]) => void | Promise<void>
  renderItem: (item: T, props: SortableItemRenderProps) => React.ReactNode
  disabled?: boolean
  // Drag-and-drop の DndContext は呼び出し元（SortableDndProvider）で提供する。
  // ここでは SortableContext / useSortable だけを行う（tbody の DOM 事故を防ぐ）。
  withDndContext?: boolean
}

export interface SortableTbodyProps<T> {
  items: T[]
  itemId: (item: T) => string
  renderItem: (item: T, props: SortableItemRenderProps) => React.ReactNode
  disabled?: boolean
  tbodyClassName?: string
}

function SortableItem<T>({
  item,
  itemId,
  renderItem,
  disabled,
}: {
  item: T
  itemId: (item: T) => string
  renderItem: (item: T, props: SortableItemRenderProps) => React.ReactNode
  disabled?: boolean
}) {
  const id = itemId(item)
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id, disabled })

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const handleProps = disabled ? {} : { ...attributes, ...listeners }

  return <>{renderItem(item, { handleProps, setNodeRef, style })}</>
}

export function SortableDndProvider<T>({
  items,
  itemId,
  onReorder,
  disabled = false,
  children,
}: SortableDndProviderProps<T>) {
  const ids = items.map((i) => itemId(i))

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const handleDragEnd = (event: DragEndEvent) => {
    if (disabled) return
    const { active, over } = event
    if (!over || active.id === over.id) return

    const oldIndex = ids.indexOf(String(active.id))
    const newIndex = ids.indexOf(String(over.id))
    if (oldIndex === -1 || newIndex === -1) return

    const newIds = arrayMove(ids, oldIndex, newIndex)
    onReorder(newIds)
  }

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      {children}
    </DndContext>
  )
}

export function SortableList<T>({
  items,
  itemId,
  onReorder,
  renderItem,
  disabled = false,
}: SortableListProps<T>) {
  const ids = items.map((i) => itemId(i))

  const sortable = (
    <SortableContext items={ids} strategy={verticalListSortingStrategy} disabled={disabled}>
      {items.map((item) => (
        <SortableItem
          key={itemId(item)}
          item={item}
          itemId={itemId}
          renderItem={renderItem}
          disabled={disabled}
        />
      ))}
    </SortableContext>
  )
  // onReorder は SortableDndProvider 側で利用されるため、ここでは使わない。
  void onReorder
  return sortable
}

// テーブル（tbody）用途: <tbody> の直下に不正要素が入らないよう、tbody自体を返す。
export function SortableTbody<T>({
  items,
  itemId,
  renderItem,
  disabled = false,
  tbodyClassName,
}: SortableTbodyProps<T>) {
  const ids = items.map((i) => itemId(i))

  return (
    <SortableContext items={ids} strategy={verticalListSortingStrategy} disabled={disabled}>
      <tbody data-sortable-tbody="true" className={tbodyClassName}>
        {items.map((item) => (
          <SortableItem key={itemId(item)} item={item} itemId={itemId} renderItem={renderItem} disabled={disabled} />
        ))}
      </tbody>
    </SortableContext>
  )
}

export function DragHandle({ handleProps, className }: { handleProps: React.HTMLAttributes<HTMLDivElement>; className?: string }) {
  return (
    <div
      {...handleProps}
      className={`cursor-grab active:cursor-grabbing touch-none p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded ${className ?? ''}`}
      title="ドラッグして並び替え"
    >
      <GripVertical className="w-4 h-4" />
    </div>
  )
}
