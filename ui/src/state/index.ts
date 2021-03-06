import React from 'react'

export type Status = 'active' | 'disconnected' | 'editing' | 'error' | 'idle'

export interface State {
    complete?: boolean
    completedAt?: string
    error?: any
    fitness?: number
    generation: number
    jobID?: number
    nextTargetImage?: string
    numWorkers?: number
    output?: string
    palette?: string
    startedAt?: string
    status: Status
    targetImage?: string
    targetImageEdges?: string
    tasks?: Record<string, any>[]
}

export const initialState: State = {
    generation: 0,
    status: 'idle',
}

export type ActionType = 'clearTarget' | 'setTarget' | 'start' | 'status' | 'update'

export interface Action {
    type: ActionType
    payload?: Record<string, any>
}

export interface StateContextType {
    dispatch: (action: Action) => void
    state: State
}

export const StateContext = React.createContext<StateContextType>({
    dispatch: (_: Action) => null,
    state: initialState,
})
