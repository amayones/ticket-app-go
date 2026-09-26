// Logika penyusunan matriks akses menu (dipisah dari komponen agar bisa
// dipakai ulang/diuji tanpa JSX).
//
// Menu PARENT = header buka-tutup tanpa halaman sendiri, jadi TIDAK punya
// centang akses: hanya menu CHILD yang dicentang. Header ikut tampil otomatis
// di sidebar bila minimal satu anaknya dicentang — dan yang tampil hanya menu
// CHILD yang dicentang itu.
export const PARENT_KIND = 'PARENT'

export function isParent(perm) {
  return perm.kind === PARENT_KIND
}

// Susun permission per modul: PARENT jadi header (tanpa centang), tiap CHILD
// masuk ke section parent-nya. Backend sudah mengirim urutan "per section"
// sehingga anak menyusul parent-nya; anak yang parent-nya tidak ada di daftar
// (mis. data lama) tetap tampil sebagai item mandiri.
// Hasil: [{ group, sections: [{ code, parent, children }], children }] —
// `children` = semua menu CHILD modul itu (dipakai untuk Pilih/Hapus semua).
export function groupPermissions(perms = []) {
  const byModule = new Map()
  for (const p of perms) {
    const key = `MODULE ${p.group || 'OTHER'}`
    if (!byModule.has(key)) byModule.set(key, [])
    byModule.get(key).push(p)
  }
  const groups = []
  for (const [group, list] of byModule) {
    const sections = []
    const sectionByCode = new Map()
    for (const p of list) {
      if (!isParent(p)) continue
      const section = { code: p.code, parent: p, children: [] }
      sectionByCode.set(p.code, section)
      sections.push(section)
    }
    for (const p of list) {
      if (isParent(p)) continue
      const section = p.parent_code ? sectionByCode.get(p.parent_code) : null
      if (section) section.children.push(p)
      else sections.push({ code: p.code, parent: null, children: [p] })
    }
    groups.push({ group, sections, children: list.filter((p) => !isParent(p)) })
  }
  return groups
}
