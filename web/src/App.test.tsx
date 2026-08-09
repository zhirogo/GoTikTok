import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from './App'

describe('App', () => {
  it('渲染 GoTikTok 占位标题', () => {
    render(<App />)
    expect(screen.getByRole('heading', { name: 'GoTikTok' })).toBeInTheDocument()
  })
})
