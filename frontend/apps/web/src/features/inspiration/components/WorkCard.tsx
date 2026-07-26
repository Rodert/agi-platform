import { HeartFilled, HeartOutlined, StarFilled, StarOutlined } from '@ant-design/icons'
import type { Work } from '../../../types'
import { useApp } from '../../../store/AppStore'
import { useNavigate } from 'react-router-dom'

export function WorkCard({ work, onOpen }: { work: Work; onOpen: (w: Work) => void }) {
  const { toggleLike, toggleCollect } = useApp()
  const nav = useNavigate()
  const media = work.type === 'video' ? work.video_url || '' : work.image_url || ''
  const createSame = () => nav(`/?prompt=${encodeURIComponent(work.prompt)}&mode=${work.type}`)

  return <article className="work-card" onClick={() => onOpen(work)}>
    {work.type === 'video' ? <video src={media} poster={work.image_url} muted preload="metadata" /> : <img src={media} alt={work.title} />}
    {work.type === 'video' && <span className="absolute right-2 top-2 rounded bg-black/60 px-2 py-1 text-xs">▶ 视频</span>}
    <div className="work-overlay">
      <b>{work.title}</b>
      <div className="work-actions">
        <button onClick={event => { event.stopPropagation(); createSame() }}>做同款</button>
        <button onClick={event => { event.stopPropagation(); nav(`/create?prompt=${encodeURIComponent(work.prompt)}`) }}>用作参考图</button>
      </div>
      <div className="work-meta">
        <span>@{work.user?.name || '用户'}</span>
        <span className="flex gap-3">
          <button aria-label="收藏" className="border-0 bg-transparent p-0" onClick={event => { event.stopPropagation(); void toggleCollect(work.id) }}>{work.is_collected ? <StarFilled className="text-[#f1c75b]" /> : <StarOutlined />}</button>
          <button aria-label="点赞" className="border-0 bg-transparent p-0" onClick={event => { event.stopPropagation(); void toggleLike(work.id) }}>{work.is_liked ? <HeartFilled className="text-[#ef7698]" /> : <HeartOutlined />} {work.likes_count}</button>
        </span>
      </div>
    </div>
  </article>
}
