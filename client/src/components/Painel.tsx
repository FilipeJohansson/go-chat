import classNames from "classnames";
import { ReactNode } from "react";

interface PainelProps {
  children: ReactNode,
  className?: string
}

export function Painel({ children, className }: PainelProps) {
  return (
    <div className={classNames(className, "gap-1 p-1.5 bg-zinc-100 bg-opacity-30 border border-zinc-100 border-opacity-50 rounded-md")}>
      {children}
    </div>
  )
}