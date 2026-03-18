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

export interface SortableListProps<T> {
  items: T[]
  itemId: (item: T) => string
  onReorder: (ids: string[]) => void | Promise<void>
  renderItem: (item: T, handleProps: React.HTMLAttributes<HTMLDivElement>) => React.ReactNode
  disabled?: boolean
}

function SortableItem<T>({
  item,
  itemId,
  renderItem,
  disabled,
}: {
  item: T
  itemId: (item: T) => string
  renderItem: (item: T, handleProps: React.HTMLAttributes<HTMLDivElement>) => React.ReactNode
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

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const handleProps = disabled ? {} : { ...attributes, ...listeners }

  return (
    <div ref={setNodeRef} style={style} className="contents">
      {renderItem(item, handleProps)}
    </div>
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
    </DndContext>
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
