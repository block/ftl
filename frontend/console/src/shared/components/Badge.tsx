export const Badge = ({ name, className }: { name: string; className?: string }) => {
  return <span className={`inline-flex items-center rounded-lg px-2 py-1 text-xs font-medium ${className}`}>{name}</span>
}
